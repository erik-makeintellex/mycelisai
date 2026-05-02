package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/internal/memory"
	"github.com/mycelis/core/internal/router"
	mycelis_nats "github.com/mycelis/core/internal/transport/nats"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

type coreRuntime struct {
	SharedDB         *sql.DB
	DBURL            string
	CogRouter        *cognitive.Router
	Guard            *governance.Guard
	MemService       *memory.Service
	NC               *nats.Conn
	ObserverNC       *nats.Conn
	ObserverNCSource string
	Router           *router.Router
	natsRuntime      *natsRuntime
}

func startCoreRuntime(ctx context.Context, natsURL string) *coreRuntime {
	dbConfig := resolveDatabaseConfig()
	dbURL := dbConfig.connectionString()

	sharedDB := openSharedDB(ctx, dbURL)
	cogRouter := loadCognitiveRouter(sharedDB)
	guard := loadGovernanceGuard()
	memService := startMemoryService(ctx, dbURL)
	natsRuntime := connectNATSLanes(natsURL)

	rt := &coreRuntime{
		SharedDB:         sharedDB,
		DBURL:            dbURL,
		CogRouter:        cogRouter,
		Guard:            guard,
		MemService:       memService,
		NC:               natsRuntime.NC,
		ObserverNC:       natsRuntime.ObserverNC,
		ObserverNCSource: natsRuntime.ObserverNCSource,
		natsRuntime:      natsRuntime,
	}

	rt.Router = startRouter(rt.ObserverNC, guard)
	startMemorySubscriber(memService, rt.ObserverNC, rt.ObserverNCSource)
	return rt
}

func (rt *coreRuntime) DrainNATS() {
	if rt != nil && rt.natsRuntime != nil {
		rt.natsRuntime.Drain()
	}
}

func openSharedDB(ctx context.Context, dbURL string) *sql.DB {
	sharedDB, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Printf("WARN: Shared DB Init Failed: %v - running without persistence.", err)
		return nil
	}

	sharedDB.SetMaxOpenConns(10)
	sharedDB.SetMaxIdleConns(5)
	sharedDB.SetConnMaxLifetime(5 * time.Minute)

	var pingErr error
	for i := 1; i <= 45; i++ {
		pingCtx, pingCancel := context.WithTimeout(ctx, 2*time.Second)
		pingErr = sharedDB.PingContext(pingCtx)
		pingCancel()
		if pingErr == nil {
			log.Println("Shared Database Connection Active.")
			break
		}
		if i == 1 || i%10 == 0 {
			log.Printf("[db] waiting for PostgreSQL (attempt %d/45): %v", i, pingErr)
		}
		time.Sleep(2 * time.Second)
	}
	if pingErr != nil {
		log.Printf("WARN: PostgreSQL unreachable after 90s: %v - running without persistence.", pingErr)
		return nil
	}

	return sharedDB
}

func loadCognitiveRouter(sharedDB *sql.DB) *cognitive.Router {
	cogRouter, err := cognitive.NewRouter("config/cognitive.yaml", sharedDB)
	if err != nil {
		log.Printf("WARN: Cognitive Config not loaded: %v. Cognitive Engine Disabled.", err)
		return nil
	}
	log.Println("Cognitive Engine Active.")
	return cogRouter
}

func loadGovernanceGuard() *governance.Guard {
	guard, err := governance.NewGuard("config/policy.yaml")
	if err != nil {
		log.Printf("WARN: Governance Policy not loaded: %v. Allowing all.", err)
		return nil
	}
	log.Println("Governance Guard Active.")
	return guard
}

func startMemoryService(ctx context.Context, dbURL string) *memory.Service {
	memService, err := memory.NewService(dbURL)
	if err != nil {
		log.Printf("WARN: Memory System Failed: %v. Continuing without persistence.", err)
		return nil
	}
	log.Println("Cortex Memory Connected.")
	go memService.Start(ctx)
	return memService
}

func startRouter(observerNC *nats.Conn, guard *governance.Guard) *router.Router {
	if observerNC == nil {
		log.Println("WARN: Router disabled (no NATS connection).")
		return nil
	}

	r := router.NewRouter(observerNC, guard)
	if err := r.Start(); err != nil {
		log.Printf("WARN: Failed to start Router: %v", err)
	}
	return r
}

func startMemorySubscriber(memService *memory.Service, observerNC *nats.Conn, observerNCSource string) {
	if memService == nil || observerNC == nil {
		return
	}

	_, err := observerNC.Subscribe(protocol.TopicSwarmWild, func(msg *nats.Msg) {
		if logEntry := buildMemoryLogEntryFromMessage(msg.Subject, msg.Data); logEntry != nil {
			memService.Push(logEntry)
		}
	})
	if err != nil {
		log.Printf("Failed to subscribe Memory Service: %v", err)
	} else {
		log.Printf("[nats] memory subscriber attached to %s lane", observerNCSource)
	}
}

type natsRuntime struct {
	NC               *nats.Conn
	ObserverNC       *nats.Conn
	ObserverNCSource string
	coreWrapper      *mycelis_nats.Client
	observerWrapper  *mycelis_nats.Client
}

func connectNATSLanes(natsURL string) *natsRuntime {
	rt := &natsRuntime{}
	ncWrapper, connErr := connectNATSWithRetry(natsURL, "Mycelis Core")
	if connErr != nil {
		log.Printf("WARN: NATS unreachable after 90s: %v. Running in DEGRADED mode (no messaging).", connErr)
		return rt
	}

	rt.coreWrapper = ncWrapper
	rt.NC = ncWrapper.Conn
	log.Printf("[nats] connected to %s", rt.NC.ConnectedUrl())

	observerWrapper, observerErr := connectNATSWithRetry(natsURL, "Mycelis Observer")
	if observerErr != nil {
		log.Printf("WARN: observer NATS lane unavailable: %v. Falling back to the primary NATS connection for router and memory fanout.", observerErr)
		rt.ObserverNC = rt.NC
		rt.ObserverNCSource = "primary"
		return rt
	}

	rt.observerWrapper = observerWrapper
	rt.ObserverNC = observerWrapper.Conn
	rt.ObserverNCSource = "observer"
	log.Printf("[nats] observer lane connected to %s", rt.ObserverNC.ConnectedUrl())
	return rt
}

func (rt *natsRuntime) Drain() {
	if rt == nil {
		return
	}
	if rt.observerWrapper != nil {
		_ = rt.observerWrapper.Drain()
	}
	if rt.coreWrapper != nil {
		_ = rt.coreWrapper.Drain()
	}
}

func connectNATSWithRetry(url, connectionName string) (*mycelis_nats.Client, error) {
	var (
		client *mycelis_nats.Client
		err    error
	)
	for i := 1; i <= 45; i++ {
		client, err = mycelis_nats.ConnectAs(url, connectionName)
		if err == nil {
			return client, nil
		}
		if i == 1 || i%10 == 0 {
			log.Printf("[nats] waiting for %s at %s (attempt %d/45): %v", connectionName, url, i, err)
		}
		time.Sleep(2 * time.Second)
	}
	return nil, err
}
