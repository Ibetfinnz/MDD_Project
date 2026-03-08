```mermaid
graph LR
    C[Client / Frontend]
    GW[Gateway Service\n:8080]

    US[User Service\n:8081]
    RS[Room Service\n:8082]
    MS[Meter Service\n:8083]
    BS[Bill Service\n:8084]

    UDB[SQLite: user.db]
    RDB[SQLite: room.db]
    MDB[SQLite: meter.db]
    BDB[SQLite: bill.db]

    RQ[RabbitMQ]
    QW[meter.water.created]
    QE[meter.electric.created]

    %% Client -> Gateway
    C -->|HTTP /api/* + JWT| GW

    %% Gateway routing
    GW -->|/api/users/*| US
    GW -->|/api/rooms/*| RS
    GW -->|/api/meters/*| MS
    GW -->|/api/bills/*| BS

    %% Auth flow
    C -->|POST /api/users/login| GW
    GW --> US
    US -->|JWT Token| C

    %% DB connections
    US --> UDB
    RS --> RDB
    MS --> MDB
    BS --> BDB

    %% Meter -> RabbitMQ -> Bill
    MS -->|publish water/electric created| RQ
    RQ --> QW
    RQ --> QE
    QW -->|consume| BS
    QE -->|consume| BS

    %% Bill Service calling other services directly (service-to-service)
    BS -->|HTTP :8082/{room_id} (fetch room)| RS
    BS -->|HTTP :8083/water/{room_id} (latest water)| MS
    BS -->|HTTP :8083/electric/{room_id} (latest electric)| MS
```