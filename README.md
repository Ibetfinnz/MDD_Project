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

1. User Service (Auth & Identity)
Base URL: http://localhost:8080/api/users

Login (เข้าสู่ระบบ):

Method: POST

URL: /login

Body (JSON): ```json
{ "username": "admin", "password": "1234" }

Check Role (ตรวจสอบสถานะปัจจุบัน):

Method: GET

URL: /check-role

Logout (ออกจากระบบ):

Method: POST

URL: /logout

2. Room Service (จัดการห้องพัก)
Base URL: http://localhost:8080/api/rooms

หมายเหตุ: Service นี้มีการเช็ค Role ดังนั้นคุณต้องทำการ Login ผ่าน User Service ก่อน

ดูห้องพักทั้งหมด:

Method: GET

URL: /rooms

เพิ่มห้องใหม่ (เฉพาะ Admin):

Method: POST

URL: /rooms

Body (JSON): ```json
{ "room_number": "301", "price": 5000, "status": "Available" }

เพิ่มผู้เช่าเข้าห้อง (เฉพาะ Admin):

Method: POST

URL: /:id/tenant (เช่น /1/tenant)

Body (JSON): { "tenant_name": "Somchai" }

3. Meter Service (บันทึกมิเตอร์น้ำ-ไฟ)
Base URL: http://localhost:8080/api/meters

จดมิเตอร์น้ำ:

Method: POST

URL: /meter/water

Body (JSON): { "room_id": "101", "unit": 15.5 }

จดมิเตอร์ไฟ:

Method: POST

URL: /meter/electric

Body (JSON): { "room_id": "101", "unit": 120.0 }

ดูประวัติมิเตอร์ไฟ (รายห้อง):

Method: GET

URL: /meter/electric/101

4. Bill Service (สรุปค่าใช้จ่าย)
Base URL: http://localhost:8080/api/bills

สร้างบิล (คำนวณอัตโนมัติ): * Service จะไปดึงราคาห้องจาก Room Service และหน่วยน้ำไฟจาก Meter Service มาคำนวณให้เอง

Method: POST

URL: /Bill/:room_id (เช่น /Bill/101)

Body (JSON): {"status": "Unpaid"}  

ดูบิลล่าสุดของห้อง:

Method: GET

URL: /Bill/101

อัปเดตสถานะการจ่ายเงิน:

Method: PATCH

URL: /Bill/101

Body (JSON): { "status": "Paid" }

## 🛠️ Technology Stack

* **Language:** Go 1.x
* **Web Framework:** Gin Gonic
* **Database:** SQLite (GORM)
* **Containerization:** Docker & Docker Compose
* **Gateway:** Reverse Proxy with Go Standard Library
