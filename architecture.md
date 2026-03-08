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

    %% Meter Service -> RabbitMQ
    MS --> RQ
    MS --> QW
    MS --> QE

    %% RabbitMQ -> Bill Service
    RQ --> BS
    QW --> BS
    QE --> BS
```

```mermaid
sequenceDiagram
    participant C as Client
    participant GW as Gateway
    participant US as UserService
    participant RS as RoomService
    participant MS as MeterService
    participant MQ as RabbitMQ
    participant BS as BillService

    Note over C,GW: Login (ครั้งเดียว)
    C->>GW: POST /api/users/login
    GW->>US: POST /login
    US-->>C: JWT token

    Note over C,MS: บันทึกมิเตอร์น้ำ/ไฟ
    C->>GW: POST /api/meters/water
    GW->>MS: POST /water (มี JWT ผ่าน Gateway)
    MS->>MS: บันทึกลง meter.db
    MS->>MQ: publish meter.water.created
    MQ-->>BS: ส่ง message ไปยัง BillService
    BS->>BS: อ่าน event และประมวลผล (ปัจจุบัน log ไว้)

    Note over C,BS: สร้างบิลค่าเช่า + ค่าน้ำไฟ
    C->>GW: POST /api/bills
    GW->>BS: POST / (สร้างบิล)
    BS->>RS: GET ห้องจาก RoomService
    RS-->>BS: ข้อมูลห้องและราคาเช่า
    BS->>MS: GET water meter ล่าสุด
    MS-->>BS: หน่วยน้ำล่าสุด
    BS->>MS: GET electric meter ล่าสุด
    MS-->>BS: หน่วยไฟล่าสุด
    BS->>BS: คำนวณยอดบิลและบันทึกลง bill.db
    BS-->>C: ส่งข้อมูลบิลกลับ

```

```mermaid
classDiagram
    class User {
      +string Username
      +string Password
      +string Role
    }

    class Room {
      +string RoomNumber
      +float Price
      +string Status
      +string TenantName
    }

    class WaterMeter {
      +string RoomID
      +float Unit
      +string Month
    }

    class ElectricMeter {
      +string RoomID
      +float Unit
      +string Month
    }

    class Bill {
      +string RoomID
      +float RentPrice
      +float WaterPrice
      +float ElectricPrice
      +float Total
      +string Month
      +string Status
    }

    Room "1" --> "*" WaterMeter : usage
    Room "1" --> "*" ElectricMeter : usage
    Room "1" --> "*" Bill : billed for
    User "1" --> "*" Room : tenant (by name/role)
```

```mermaid
sequenceDiagram
    participant MS as MeterService
    participant RQ as RabbitMQ
    participant BS as BillService

    Note over MS,RQ: Startup MeterService
    MS->>RQ: connect (RABBITMQ_URL)
    MS->>RQ: declare queues\nmeter.water.created, meter.electric.created

    Note over BS,RQ: Startup BillService
    BS->>RQ: connect (RABBITMQ_URL)
    BS->>RQ: declare queues
    BS->>RQ: start consumers\nfor both queues

    Note over MS,RQ: When water meter is created
    MS->>RQ: publish message to meter.water.created
    RQ-->>BS: deliver message to consumer
    BS->>BS: handle event (ตอนนี้แค่ log)

    Note over MS,RQ: When electric meter is created
    MS->>RQ: publish message to meter.electric.created
    RQ-->>BS: deliver message to consumer
    BS->>BS: handle event (ตอนนี้แค่ log)