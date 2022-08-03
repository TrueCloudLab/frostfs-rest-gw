
# frostfs-rest-gw

FrostFS REST Gateway bridges FrostFS internal protocol and REST API server.

### Open API specification

See full [API spec](/v1/docs).

### Basic concept

Using this API you can interact with FrostFS nodes and manage containers and objects.

#### Container

To create container you must provide `PlacementPolicy` and `BasicACL`.

##### Placement policy

Placement policy allows you control where and how container (and its object) is stored.
For example, you want to store 3 copy of every object, so you can use the following policy:

```
REP 3
```

[More about policy](https://github.com/nspcc-dev/neofs-spec/blob/7ae698ebbe68c689cab2aba518312e7d3eea403c/01-arch/02-policy.md).

##### Basic ACL

Basic ACL is a part of the container structure, and it is always created simultaneously with the container.
Therefore, it is never subject to any changes. It is a 32-bit integer with a bit field in the following format:

<img src="docs/acl-basic.svg" alt="acl-basic">

| Symbol | Meaning | Description                                                                                    |
|--------|:--------|------------------------------------------------------------------------------------------------|
| **B**  | Bearer  | Allows using Bear Token ACL rules to replace eACL rules                                        |
| **U**  | User    | The owner of the container identified by the public key linked to the container                |
| **S**  | System  | Inner Ring and/or container nodes in the current version of network map                        |
|        |         | IR nodes can only perform `GetRangeHash`, `Head`, and `Search` necessary for data audit.       |
|        |         | Container nodes can only do things required for the replication.                               |
| **O**  | Others  | Clients that do not match any of the categories above                                          |
| **F**  | Final   | Flag denying Extended ACL. If set, Basic ACL check is final, Extended ACL is ignored           |
| **X**  | Sticky  | Flag denying different owners of the request and the object                                    |
|        |         | If set, object in `Put` request must have one `Owner` and be signed with the same signature    |
|        |         | If not set, the object must be correct but can be of any owner.                                |
|        |         | The nodes falling for `SYSTEM` role are exception from this rule. For them the bit is ignored. |
| **0**  | Deny    | Denies operation of the identified category                                                    |
| **1**  | Allow   | Allows operation of the identified category                                                    |

To upload objects with bearer token your container must have Bearer bits set. 
For example, you can use `0x0FBFBFFF` or predefined `eacl-public-read-write` values. 

Also don't forget set appropriate eACL to restrict your container.

[More about ACL](https://github.com/nspcc-dev/neofs-spec/blob/4f8d945dfbd2a313ebd406746cf38b9de9da6038/01-arch/07-acl.md).

#### Object
To create object you must provide `containerId` and `fileName`.
Additionally, you can provide `payload` (base64 encoded data) and `attributes`.

Attribute is key value data that is stored with object. Key and value must be in utf8 format and must not be empty.

Valid attribute:
* `MyAttribute: 'some value'`

Invalid attribute:
* `MyAttribute: ''`

Also, you can use this attribute to further object searching.

### Status codes

More about FrostFS status code you can
find [here](https://github.com/nspcc-dev/neofs-spec/blob/master/20-api-v2/status.md).

### Storage groups

The concept of a storage group has been introduced to reduce the dependence of the complexity of
the check on the number of stored objects in the system.

The consistency and availability of multiple objects on the network are achieved by validating the
storage group without saving meta information and performing validation on each object.

`StorageGroup` keeps verification information for Data Audit sessions. Objects that require paid storage
guaranties are gathered in `StorageGroups` with additional information used for proof of storage
checks. A `StorageGroup` can be created only for objects from the same container.

A `StorageGroup` are objects of a special type with the payload containing the serialized protobuf
structure. For the details on the format please refer to the [API specification](https://github.com/nspcc-dev/neofs-spec/blob/master/20-api-v2/storagegroup.md) in the corresponding section.

StorageGroup structure has information about:
* Total size of the payloads of objects in the storage group
* Homomorphic hash from the concatenation of the payloads of the storage group members. The order of concatenation is the same as the order of the members in the members field.

* Last NeoFS epoch number of the storage group lifetime
* Alpha-numerically sorted list of member objects
