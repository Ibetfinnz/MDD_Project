# MDD_Project
---

```markdown
# 🏢 Dormitory Management System (Microservices Backend)
โปรเจกต์ระบบบริหารจัดการหอพัก พัฒนาด้วยสถาปัตยกรรม Microservices โดยใช้ภาษา **Go (Gin Framework)** และ **SQLite** เป็นฐานข้อมูลแยกแต่ละบริการ

## 🏗️ System Architecture
ระบบประกอบด้วย 5 บริการหลักที่ทำงานเชื่อมต่อกันผ่าน API Gateway:
1. **Gateway Service (Port 8080):** ประตูหน้าด่าน (Reverse Proxy) รวม API ทุกตัว ตรวจสอบ JWT แล้วส่งต่อข้อมูลผู้ใช้ให้แต่ละ Service
2. **User Service (Port 8081):** จัดการข้อมูลผู้ใช้งาน (Admin/Staff) และระบบ Login
3. **Room Service (Port 8082):** จัดการข้อมูลห้องพัก สถานะห้อง และข้อมูลผู้เช่า
4. **Meter Service (Port 8083):** จัดการจดบันทึกมิเตอร์น้ำและไฟรายเดือน
5. **Bill Service (Port 8084):** สรุปยอดค่าใช้จ่าย คำนวณบิล และจัดการสถานะการชำระเงิน

---

## 🚀 วิธีติดตั้งและรันโปรเจกต์ (Getting Started)

### 1. สิ่งที่ต้องมีในเครื่อง (Prerequisites)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/) (ติดตั้งและเปิดให้สถานะเป็น Running)
- [Postman](https://www.postman.com/downloads/) (สำหรับทดสอบ API)

### 2. ขั้นตอนการติดตั้ง
1. **Clone Repository:**
   ```bash
   git clone (https://github.com/Ibetfinnz/MDD_Project.git)
   cd Project_MDD

```

2. **รันระบบด้วย Docker Compose:**
ใช้คำสั่งนี้เพื่อ Build และรันทุก Microservices ขึ้นมาพร้อมกัน:
```bash
docker-compose up --build

```


*หมายเหตุ: หากเจอ Error เรื่องชื่อ Container ซ้ำ ให้ลบของเก่าออกก่อนด้วยคำสั่ง `docker container prune -f*`
3. **ตรวจสอบสถานะ:**
ทุก Service จะต้องขึ้นสถานะ `Running` และ Gateway จะเปิดรอรับคำสั่งที่พอร์ต `8080`

---

## 📮 รายละเอียดการทดสอบ API ด้วย Postman

**สำคัญ:** ทุก Request จะต้องยิงไปที่พอร์ต `8080` (Gateway) เท่านั้น

## 🚀 API Documentation (Microservices via Gateway)

**Base URL:** `http://localhost:8080` (API Gateway)

### 1. User Service (`/api/users`)

จัดการเรื่องการยืนยันตัวตนและข้อมูลผู้ใช้งาน (Login + JWT)

| Feature | Method | Endpoint | Description / Role |
| --- | --- | --- | --- |
| **Login** | `POST` | `/login` | เข้าสู่ระบบและรับโทเคนสำหรับใช้งาน API |
| **Get Current User** | `GET` | `/me` | ดูข้อมูลผู้ใช้ที่กำลังล็อกอินอยู่ |
| **Get All Users** | `GET` | `/users` | **Admin** ดูรายชื่อผู้ใช้ทั้งหมด |

**Request Body ตัวอย่าง**

- Login
   ```json
   {
      "username": "admin",
      "password": "1234"
   }
   ```

#### 🔐 วิธีใช้ Token ตอนทดสอบ (Postman)

1. เริ่มจาก Login ก่อน
    - Method: `POST`
    - URL: `http://localhost:8080/api/users/login`
    - Body (JSON):
       ```json
       {
          "username": "admin",
          "password": "1234"
       }
       ```
    - ถ้า Login สำเร็จ จะได้ response ประมาณนี้:
       ```json
       {
          "message": "Login successful",
          "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
       }
       ```

2. คัดลอกค่าในฟิลด์ `token`

3. เวลาเรียก API ที่ต้องใช้สิทธิ์ (เช่น `/api/users/me`, `/api/rooms/rooms` ฯลฯ)
   - ไปที่แท็บ **Authorization** ใน Postman
   - เลือก Type = `Bearer Token`
   - วางค่าที่คัดลอกจากฟิลด์ `token` ลงในช่อง **Token**

เมื่อกด Send แล้ว Postman จะสร้าง Header `Authorization: Bearer <token>` ให้อัตโนมัติ และ API ที่อยู่หลัง middleware JWT จะสามารถอ่าน username / role จาก token และตอบข้อมูลให้ตามสิทธิ์ได้

---

### 2. Room Service (`/api/rooms`)

จัดการข้อมูลห้องพัก (ต้อง Login ก่อนใช้งาน ผ่าน Gateway)

| Feature | Method | Endpoint | Description / Role |
| --- | --- | --- | --- |
| **Get All Rooms** | `GET` | `/api/rooms/` | ดูรายการห้องทั้งหมด |
| **Get Room Detail** | `GET` | `/api/rooms/:id` | ดูรายละเอียดห้องตาม ID |
| **Create Room** | `POST` | `/api/rooms/` | **Admin** สร้างห้องใหม่ |
| **Update Room** | `PATCH` | `/api/rooms/:id` | **Admin** แก้ไขข้อมูลห้อง |
| **Add Tenant** | `POST` | `/api/rooms/:id/tenant` | **Admin** เพิ่มผู้เช่าลงในห้อง |

**Request Body ตัวอย่าง**

- Create Room
   ```json
   {
      "room_number": "301",
      "price": 5000,
      "status": "Available"
   }
   ```

- Update Room (ตัวอย่างแก้เฉพาะราคา)
   ```json
   {
      "price": 5500
   }
   ```

- Add Tenant
   ```json
   {
      "tenant_name": "Somchai"
   }
   ```

---

### 3. Meter Service (`/api/meters`)

บันทึกและดูประวัติการใช้ค่าน้ำ-ค่าไฟ (ต้อง Login ผ่าน Gateway)

| Feature | Method | Endpoint | Description / Role |
| --- | --- | --- | --- |
| **Record Water** | `POST` | `/api/meters/water` | **Admin** บันทึกค่ามิเตอร์น้ำ |
| **Record Electric** | `POST` | `/api/meters/electric` | **Admin** บันทึกค่ามิเตอร์ไฟ |
| **Water History** | `GET` | `/api/meters/water/:room_id` | ดูประวัติการใช้น้ำของห้อง |
| **Electric History** | `GET` | `/api/meters/electric/:room_id` | ดูประวัติการใช้ไฟของห้อง |

**Request Body ตัวอย่าง**

- Record Water
   ```json
   {
      "room_id": "101",
      "unit": 15.5
   }
   ```

- Record Electric
   ```json
   {
      "room_id": "101",
      "unit": 120.0
   }
   ```

---

### 4. Bill Service (`/api/bills`)

สรุปค่าใช้จ่ายประจำเดือน (ดึงข้อมูลจาก Room และ Meter Service ผ่าน Gateway)

| Feature | Method | Endpoint | Description / Role |
| --- | --- | --- | --- |
| **Create Bill** | `POST` | `/api/bills/:room_id` | **Admin** สร้างบิลค่าใช้จ่ายห้อง |
| **Get Latest Bill** | `GET` | `/api/bills/:room_id` | ดูบิลล่าสุดของห้อง |
| **Get All Bills** | `GET` | `/api/bills/` | **Admin** ดูรายการบิลทั้งหมด |
| **Update Status** | `PATCH` | `/api/bills/:room_id` | **Admin** แก้ไขข้อมูลหรือสถานะบิล |

**Request Body ตัวอย่าง**

- Create Bill
   ```json
   {
      "status": "Unpaid"
   }
   ```

- Update Status (ตัวอย่างแก้ราคาค่าน้ำ หรือฟิลด์อื่น ๆ)
   ```json
   {
      "water_price": 200
   }
   ```

## 🛠️ Technology Stack

* **Language:** Go 1.x
* **Web Framework:** Gin Gonic
* **Database:** SQLite (GORM)
* **Containerization:** Docker & Docker Compose
* **Gateway:** Reverse Proxy with Go Standard Library
