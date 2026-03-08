graph LR
    subgraph ClientSide[Client]
        C[Client / Frontend]
    end

    subgraph GatewayLayer[API Gateway]
        GW[Gateway Service\n:8080\n- JWT validation\n- Circuit Breaker\n- Proxy /api/*]
    end

    subgraph Services[Microservices]
        US[User Service\n:8081\n- Login\n- JWT Issuer]
        RS[Room Service\n:8082\n- Rooms\n- Tenant]
        MS[Meter Service\n:8083\n- Water/Electric Meters]
        BS[Bill Service\n:8084\n- Rent & Utility Bills]
    end

    subgraph Databases[Databases (SQLite)]
        UDB[(user.db)]
        RDB[(room.db)]
        MDB[(meter.db)]
        BDB[(bill.db)]
    end

    subgraph MQ[RabbitMQ]
        RQ[(RabbitMQ\n5672 / 15672)]
        QW[[meter.water.created]]
        QE[[meter.electric.created]]
    end

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
    MS -->|publish\nwater/electric created| RQ
    RQ --> QW
    RQ --> QE
    QW -->|consume| BS
    QE -->|consume| BS

    %% Bill Service calling other services directly (service-to-service)
    BS -->|HTTP :8082/{room_id}\n(fetch room)| RS
    BS -->|HTTP :8083/water/{room_id}\n(fetch latest water)| MS
    BS -->|HTTP :8083/electric/{room_id}\n(fetch latest electric)| MS