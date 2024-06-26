## AcceptLND

AcceptLND is a channel requests management tool based on policies for LND.

## Usage

```bash
acceptlnd [-config CONFIG] [-debug] [-version]

Parameters:
  -config          Path to the configuration file (default: "acceptlnd.yml")
  -debug           Enable debug level logging
  -version         Print the current version
```

## Installation

Download the binary from the [Releases](https://github.com/aftermath2/acceptlnd/releases) page, use docker or compile it yourself.

<details><summary>Docker</summary>

```console
docker build -t acceptlnd .
# The configuration, certificate and macaroon must be mounted into the container.
# The paths specified in the configuration file can be absolute or relative to the mount path.
docker run --network=host -v <config_files_mount> acceptlnd <flags>
```

</details>

<details><summary>Build from source</summary>

> Requires Go 1.18+ installed

```console
git clone https://github.com/aftermath2/acceptlnd
cd acceptlnd
go build -o acceptlnd -ldflags="-s -w" .
```

</details>

## Configuration

The configuration file can be passed as a flag (`-config="<path>"`) when executing the binary, the default value is `acceptlnd.yml`.

Configuration schema:

| Key | Type | Required | Description |
| -- | -- | -- | -- |
| **rpc_address** | string | 🗸 | LND GRPC address (`host:port`) |
| **certificate_path** | string | 🗸 | Path to LND's TLS certificate |
| **macaroon_path** | string | 🗸 | Path to the macaroon file. See [macaroon](#macaroon) |
| **policies** | [][Policy](#policy) | X | Set of policies to enforce |

### Macaroon

AcceptLND needs a macaroon to communicate with the LND instance to manage channel requests.

Although `admin.macaroon` can be used, it is recommended baking a fine-grained macaroon that gives AcceptLND access just to the RPC methods it uses. To bake it, execute:

```
lncli bakemacaroon uri:/lnrpc.Lightning/ChannelAcceptor uri:/lnrpc.Lightning/GetInfo uri:/lnrpc.Lightning/GetNodeInfo --save_to acceptlnd.macaroon
```

Once created, specify its path in the `macaroon_path` field of the configuration file, it can be relative or absolute.

## Policy

Policies define a set of requirements that must be met for a request to be accepted. A configuration may have an unlimited number of policies, they are evaluated from top to bottom.

A policy would only be enforced if its conditions are satisfied, or if it has no conditions.

| Key | Type | Description |
| -- | -- | -- |
| **conditions** | [Conditions](#conditions) | Set of conditions that must be met to enforce the policies |
| **reject_all** | boolean | Reject all channel requests |
| **allow_list** | []string | List of nodes public keys whose requests will be accepted |
| **block_list** | []string | List of nodes public keys whose requests will be rejected |
| **accept_zero_conf_channels** | boolean | Whether to accept zero confirmation channels |
| **zero_conf_list** | []string | List of nodes public keys whose zero conf requests will be accepted. Requires `accept_zero_conf_channels` to be `true` | 
| **reject_private_channels** | boolean | Whether private channels should be rejected |
| **max_channels** | int | Maximum number of channels. Compared against the sum of the node's active, pending and inactive channels |
| **min_accept_depth** | int | Number of confirmations required before considering the channel open |
| **request** | [Request](#request) | Parameters related to the channel opening request |
| **node** | [Node](#node) | Parameters related to the channel initiator |

Here's a simple example:

```yml
policies:
  -
    conditions:
      is_private: true
    request:
      channel_capacity:
        min: 2_000_000
```

This policy only applies to private channels and will reject requests with a capacity lower than 2 million sats. 

> [!Note]
> The denomination used in all the numbers is **satoshis**.
>
> More examples can be found at [/examples](./examples/).

### Conditions

Conditions are used to evaluate policies conditionally. If they are specified, all of them must resolve to true or the policy is skipped.

They are defined in the configuration exactly the same way policies are, only a few fields change.

| Key | Type | Description |
| -- | -- | -- |
| **is** | []string | List of nodes public keys to which policies should be applied |
| **is_not** | []string | List of nodes public keys to which policies should not be applied |
| **is_private** | boolean | Match private channels |
| **wants_zero_conf** | boolean | Match zero confirmation channels |
| **request** | [Request](#request) | Parameters related to the channel opening request |
| **node** | [Node](#node) | Parameters related to the initiator node |

### Request

Parameters related to the channel opening request.

| Key | Type | Description |
| -- | -- | -- |
| **channel_capacity** | range | Requested channel size |
| **channel_reserve** | range | Requested channel reserve |
| **push_amount** | range | Pushed amount of sats |
| **csv_delay** | range | Requested CSV delay |
| **max_accepted_htlcs** | range | The total number of incoming HTLC's that the initiator will accept |
| **min_htlc** | range | The smallest HTLC in millisatoshis that the initiator will accept |
| **max_value_in_flight** | range | The maximum amount of coins in millisatoshis that can be pending in the channel |
| **dust_limit** | range | The dust limit of the initiator's commitment transaction |
| **commitment_types** | []int | Accepted channel commitment types. See [lnrpc.CommitmentTypes](https://lightning.engineering/api-docs/api/lnd/lightning/channel-acceptor/index.html#lnrpccommitmenttype) |

### Node

Parameters related to the node that is initiating the channel.

| Key | Type | Description |
| -- | -- | -- |
| **age** | range | Peer node age in blocks, based on the oldest announced channel |
| **capacity** | range | Peer node capacity |
| **hybrid** | boolean | Whether the peer will be required to be hybrid |
| **feature_flags** | []int | Feature flags the peer node must know. Check out [lnrpc.FeatureBit](https://lightning.engineering/api-docs/api/lnd/lightning/query-routes#lnrpcfeaturebit) |
| **Channels** | [Channels](#Channels) | Initiator node channels |

### Channels

Parameters related to the initiator node's channels.

| Key | Type | Description |
| -- | -- | -- |
| **number** | range | Peer's number of channels |
| **capacity** | stat_range | Channels size |
| **zero_base_fees** | boolean | Whether the peer's channels must all have zero base fees |
| **block_height** | stat_range | Channels block height |
| **time_lock_delta** | stat_range | Channels time lock delta |
| **min_htlc** | stat_range | Channels minimum HTLC |
| **max_htlc** | stat_range | Channels maximum HTLC |
| **last_update_diff** | stat_range | Channels last update difference to the time of the request (seconds) |
| **together** | range | Number of channels that the host node and initiator node have together |
| **fee_rates** | stat_range | Channels fee rates |
| **base_fees** | stat_range | Channels base fees |
| **disabled** | stat_range | Number of disabled channels. The value type is float and should be between 0 and 1 |
| **inbound_fee_rates** | stat_range | Channels inbound fee rates |
| **inbound_base_fees** | stat_range | Channels inbound base fees |
| **peers** | [Peers](#Peers) | Initiator node channels parameters on the peers' side |

> [!Note]
> **Inbound** fees were added in LND v0.18.0-beta and they represent fees for the movement of incoming funds. A positive value would discourage peers from routing to the channel and a negative value would incentivize them.

#### Peers

Initiator node channels parameters on the peers' side.

| Key | Type | Description |
| -- | -- | -- |
| **fee_rates** | stat_range | Channels fee rates |
| **base_fees** | stat_range | Channels base fees |
| **disabled** | stat_range | Number of disabled channels. The value type is float and should be between 0 and 1 |
| **inbound_fee_rates** | stat_range | Channels inbound fee rates |
| **inbound_base_fees** | stat_range | Channels inbound base fees |

#### Range

A range may have a minimum value, a maximum value or both defined. All values are in **satoshis**.

> `Min` and `Max` are inclusive, they include the value assigned: `[Min, Max]`.

##### Example

```yml
request:
  channel_capacity:
    min: 2_000_000
    max: 50_000_000
```

#### Statistic range (stat_range)

Statistic ranges work just like ranges but they compare values against the node's data set after being aggregated using an operation.

##### Example

```yml
node:
  channels:
    outgoing_fee_rates:
      operation: median
      min: 0
      max: 100
```

#### Operations

- **mean** (default): average of a list of numbers.
- **median**: middle value in a list ordered from smallest to largest.
- **mode**: most frequently occurring value on a list.
- **range**: difference between the biggest and the smallest number.
