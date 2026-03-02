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

### 1. User Service (จัดการผู้ใช้)

| Action | Method | URL | Body (JSON) |
| --- | --- | --- | --- |
| สร้าง User ใหม่ | `POST` | `http://localhost:8080/api/users/users` | `{"name": "admin", "email": "a@test.com", "password": "123", "role": "admin"}` |
| ดู User ทั้งหมด | `GET` | `http://localhost:8080/api/users/users` | - |

### 2. Room Service (จัดการห้องพัก)

| Action | Method | URL | Body (JSON) |
| --- | --- | --- | --- |
| เพิ่มห้องพักใหม่ | `POST` | `http://localhost:8080/api/rooms/rooms/` | `{"room_number": "101", "type": "VIP", "price": 5000, "status": "Available"}` |
| ดูรายชื่อห้องทั้งหมด | `GET` | `http://localhost:8080/api/rooms/rooms/` | - |
| ดูข้อมูลห้องราย ID | `GET` | `http://localhost:8080/api/rooms/rooms/1` | - |
| เพิ่มผู้เช่าเข้าห้อง | `POST` | `http://localhost:8080/api/rooms/rooms/1/tenant` | `{"tenant_name": "Finnz"}` |

### 3. Meter Service (จดมิเตอร์น้ำ-ไฟ)

| Action | Method | URL | Body (JSON) |
| --- | --- | --- | --- |
| บันทึกมิเตอร์น้ำ | `POST` | `http://localhost:8080/api/meters/meter/water` | `{"room_id": "1", "unit": 10.5}` |
| บันทึกมิเตอร์ไฟ | `POST` | `http://localhost:8080/api/meters/meter/electric` | `{"room_id": "1", "unit": 150}` |
| ดูประวัติน้ำทั้งหมด | `GET` | `http://localhost:8080/api/meters/meter/water` | - |
| ดูประวัติไฟทั้งหมด | `GET` | `http://localhost:8080/api/meters/meter/electric` | - |

### 4. Bill Service (สรุปยอดบิล)

| Action | Method | URL | Body (JSON) |
| --- | --- | --- | --- |
| สร้างบิลใหม่ | `POST` | `http://localhost:8080/api/bills/Bill/1` | `{"rent_price": 5000, "water_price": 100, "electric_price": 400}` |
| ดูบิลทั้งหมด | `GET` | `http://localhost:8080/api/bills/Bill/` | - |
| อัปเดตสถานะการจ่ายเงิน | `PATCH` | `http://localhost:8080/api/bills/Bill/1` | `{"status": "Paid"}` |

---

## 🛠️ Technology Stack

* **Language:** Go 1.x
* **Web Framework:** Gin Gonic
* **Database:** SQLite (GORM)
* **Containerization:** Docker & Docker Compose
* **Gateway:** Reverse Proxy with Go Standard Library

```

---

### คำแนะนำเพิ่มเติม:
1.  **ไฟล์ .db:** ผมแนะนำให้คุณเพิ่ม `*.db` ลงในไฟล์ `.gitignore` ด้วย เพื่อไม่ให้ไฟล์ฐานข้อมูลที่เกิดจากการรันในเครื่องคุณถูกอัปโหลดขึ้นไปบน GitHub
2.  **ชื่อ Repository:** ตอนที่คุณนำ URL ไปใส่ในส่วน `git clone` อย่าลืมแก้ให้เป็นชื่อโปรเจกต์จริงของคุณนะครับ

```
