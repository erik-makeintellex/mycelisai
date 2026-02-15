from invoke import Collection
from ops import core, k8s, proto_relay, misc, interface

ns = Collection()
ns.add_collection(core.ns)
ns.add_collection(k8s.ns)
ns.add_collection(proto_relay.ns_proto)
ns.add_collection(proto_relay.ns_relay)
# ns.add_collection(misc.ns) # Removed incorrect import
ns.add_collection(misc.ns_clean)
ns.add_collection(misc.ns_team)
ns.add_collection(interface.ns)

from ops import device, test
ns.add_collection(device.ns)
ns.add_collection(test.ns)




