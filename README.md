# MDD_Project
---

```markdown
# 🏢 Dormitory Management System (Microservices Backend)
โปรเจกต์ระบบบริหารจัดการหอพัก พัฒนาด้วยสถาปัตยกรรม Microservices โดยใช้ภาษา **Go (Gin Framework)** และ **SQLite** เป็นฐานข้อมูลแยกแต่ละบริการ

## 🏗️ System Architecture
ระบบประกอบด้วย 5 บริการหลักที่ทำงานเชื่อมต่อกันผ่าน API Gateway:
1. **Gateway Service (Port 8080):** ประตูหน้าด่าน (Reverse Proxy) ทำหน้าที่รวม API ทุกตัวและจัดการ CORS
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

เพื่อให้ข้อมูลดูเป็นระเบียบและง่ายต่อการนำไปใส่ในไฟล์ `README.md` ของโปรเจกต์บน GitHub ผมสรุปข้อมูล API ทั้งหมดในรูปแบบตารางให้ตามนี้ครับ

---

## 🚀 API Documentation (Microservices via Gateway)

**Base URL:** `http://localhost:8080` (API Gateway)

### 1. User Service (`/api/users`)

จัดการเรื่องการยืนยันตัวตนและการตรวจสอบสิทธิ์

| Feature | Method | Endpoint | Request Body (JSON) | Description |
| --- | --- | --- | --- | --- |
| **Login** | `POST` | `/login` | `{"Username" : "admin",Password: "1234"}, {"Username" : "tenant1",Password: "1234"}` | เข้าสู่ระบบเพื่อกำหนดสิทธิ์การใช้งาน |
| **Check Role** | `GET` | `/check-role` | - | ตรวจสอบข้อมูล User และ Role ที่ล็อกอินอยู่ |
| **Logout** | `POST` | `/logout` | - | ออกจากระบบ |

---

### 2. Room Service (`/api/rooms`)

จัดการข้อมูลห้องพัก (ต้อง Login ก่อนใช้งาน)

| Feature | Method | Endpoint | Request Body (JSON) | Role |
| --- | --- | --- | --- | --- |
| **Get All Rooms** | `GET` | `/rooms` | - | Any |
| **Get Room Detail** | `GET` | `/rooms/:id` | - | Any |
| **Create Room** | `POST` | `/rooms` | `{"room_number": "301", "price": 5000, "status": "Available"}` | **Admin** |
| **Update Room** | `PATCH` | `/rooms/:id` | `{"price": 5500}` | **Admin** |
| **Add Tenant** | `POST` | `/rooms/:id/tenant` | `{"tenant_name": "Somchai"}` | **Admin** |

---

### 3. Meter Service (`/api/meters`)

บันทึกและดูประวัติการใช้ค่าน้ำ-ค่าไฟ

| Feature | Method | Endpoint | Request Body (JSON) | Description |
| --- | --- | --- | --- | --- |
| **Record Water** | `POST` | `/meter/water` | `{"room_id": "101", "unit": 15.5}` | บันทึกมิเตอร์น้ำล่าสุด |
| **Record Electric** | `POST` | `/meter/electric` | `{"room_id": "101", "unit": 120.0}` | บันทึกมิเตอร์ไฟล่าสุด |
| **Water History** | `GET` | `/meter/water/:room_id` | - | ดูหน่วยน้ำล่าสุดของห้องนั้น |
| **Electric History** | `GET` | `/meter/electric/:room_id` | - | ดูหน่วยไฟล่าสุดของห้องนั้น |

---

### 4. Bill Service (`/api/bills`)

สรุปค่าใช้จ่ายประจำเดือน (ดึงข้อมูลจาก Room และ Meter Service)

| Feature | Method | Endpoint | Request Body (JSON) | Description |
| --- | --- | --- | --- | --- |
| **Create Bill** | `POST` | `/Bill/:room_id` | `{"status": "Unpaid"}` | ระบบจะคำนวณเงินรวมให้อัตโนมัติ |
| **Get Latest Bill** | `GET` | `/Bill/:room_id` | - | ดูข้อมูลบิลล่าสุดของห้อง |
| **Get All Bills** | `GET` | `/Bill/` | - | ดูรายการบิลทั้งหมดที่มีในระบบ |
| **Update Status** | `PATCH` | `/Bill/:room_id` | `{"status": "Paid"}` | อัปเดตสถานะการจ่ายเงิน |

---

> **Tip สำหรับ GitHub:** คุณสามารถก๊อปปี้ Markdown ด้านบนไปวางในไฟล์ `README.md` ได้เลยครับ ตารางจะแสดงผลอย่างสวยงามบนหน้าโปรเจกต์ของคุณ

มีส่วนไหนของตารางที่อยากให้เพิ่มรายละเอียด เช่น **HTTP Status Code** (200, 201, 401) ไหมครับ?
## 🛠️ Technology Stack

* **Language:** Go 1.x
* **Web Framework:** Gin Gonic
* **Database:** SQLite (GORM)
* **Containerization:** Docker & Docker Compose
* **Gateway:** Reverse Proxy with Go Standard Library
