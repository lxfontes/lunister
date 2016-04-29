## how does it work?

```

+------------------------+                               +------------------------+
| +-------+  +-------+   |       +---------------+       | +-------+  +-------+   |
| |       |  |       |   |       |               |       | |       |  |       |   |
| |trunk  |  |  tap  +----------->   lunister    <---------+ tap   |  |clients|   |
| |       |  |       |   |       |               |       | |       |  |       |   |
| +-------+  +-------+   |       +-------+-------+       | +-------+  +-------+   |
|                        |               |               |                        |
|                        |               |               |   client bridge        |
+------------------------+               |               +------------------------+
                                 +-------v-------+
                                 |               |
                                 | arp_manager   |
                                 |               |
                                 +---------------+

```

- Prevents clients from sending arp requests to external resources
- Simulate security groups for ipv4 tcp/udp/icmp

~~arp_manager is such a bad name~~

## Testing with vagrant

```bash
# connect to server side
ip netns exec trunkns bash

# connect to client side
ip netns exec clientns bash

```

make sure to bring up arp_manager && lunister.

you will have to recompile arp_manager with your ether addresses (trunk 10.0.0.1)

