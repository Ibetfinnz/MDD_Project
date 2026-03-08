```mermaid
graph LR
    C[Client / Frontend]
    GW[Gateway Service :8080]

    US[User Service :8081]
    RS[Room Service :8082]
    MS[Meter Service :8083]
    BS[Bill Service :8084]

    UDB[SQLite: user.db]
    RDB[SQLite: room.db]
    MDB[SQLite: meter.db]
    BDB[SQLite: bill.db]

    RQ[RabbitMQ]
    QW[meter.water.created]
    QE[meter.electric.created]

    %% Client -> Gateway
    C --> GW

    %% Gateway routing
    GW --> US
    GW --> RS
    GW --> MS
    GW --> BS

    %% Auth flow
    US --> C

    %% DB connections
    US --> UDB
    RS --> RDB
    MS --> MDB
    BS --> BDB

    %% Meter -> RabbitMQ -> Bill
    MS --> RQ
    RQ --> QW
    RQ --> QE
    QW --> BS
    QE --> BS

    %% Bill Service calling other services directly (service-to-service)
    BS --> RS
    BS --> MS