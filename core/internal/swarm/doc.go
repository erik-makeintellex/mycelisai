package swarm

// Package swarm implements the "Swarm Intelligence" layer of Mycelis.
// It creates a hierarchical multi-agent system inspired by biological neural networks.
//
// Concepts:
// - Soma (The Executive): Central orchestrator dealing with User Intent.
// - Axon (The Messenger): Signal optimizer and router (Soma's Assistant).
// - Action Cores: Teams of agents performing logical/computational work.
// - Expression Cores: Teams of agents performing output/actuation work.
//
// Bus Topology:
// - swarm.global.*: The public nervous system (Guarded).
// - swarm.team.<id>.internal.*: Private team chatter.
// - swarm.team.<id>.signal.*: Public team output.
