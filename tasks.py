from invoke import Collection
from ops import core, k8s, proto_relay, misc

ns = Collection()
ns.add_collection(core.ns, "core")
ns.add_collection(k8s.ns, "k8s")
ns.add_collection(proto_relay.ns_proto, "proto")
ns.add_collection(proto_relay.ns_relay, "relay")
ns.add_collection(misc.ns_clean, "clean")
ns.add_collection(misc.ns_team, "team")
