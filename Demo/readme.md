# คู่มือใช้ Postman สำหรับ Demo ระบบ Dormitory Management (Microservices)

เอกสารนี้อธิบายทุกอย่างที่ต้องตั้งค่าใน Postman และลำดับการยิง API ตอน Demo เพื่อให้เห็น flow ของระบบชัดเจน ผ่าน API Gateway เพียงตัวเดียว (พอร์ต 8080)

---

## 1. เตรียมระบบให้พร้อมก่อนเปิด Postman

1. เปิด Docker Desktop ให้เป็น Running
2. ไปที่โฟลเดอร์โปรเจกต์ในเครื่อง

	 ```bash
	 cd C:\Users\Aorus\Desktop\Microservice\MDD_Project
	 docker compose up --build
	 ```

3. รอให้ทุก service ขึ้นมา (user, room, meter, bill, gateway, rabbitmq)
4. เช็คเบื้องต้นด้วย browser หรือ curl:

	 - เปิด `http://localhost:8080/` ควรเห็น response จาก Gateway หรือจาก user-service (แล้วแต่ตั้งค่า)

---

## 2. สร้าง Environment ใน Postman

1. เปิด Postman
2. ที่มุมขวาบน กดปุ่ม **Environments** → **+** สร้าง Environment ใหม่
3. ตั้งชื่อเช่น `Dormitory-Microservices-Local`
4. เพิ่มตัวแปร (Variables) ดังนี้

	 | Variable          | Initial Value              | Description                                      |
	 | ----------------- | -------------------------- | ------------------------------------------------ |
	 | `base_url`        | `http://localhost:8080`    | URL ของ API Gateway                             |
	 | `admin_username`  | `admin`                    | user admin (seed มาให้แล้วใน user-service)     |
	 | `admin_password`  | `1234`                     | password ของ admin                              |
	 | `tenant_username` | `tenant1`                  | user ผู้เช่า (seed มาให้แล้ว)                  |
	 | `tenant_password` | `1234`                     | password ของ tenant1                            |
	 | `token`           | (ปล่อยว่างก่อน)           | ใช้เก็บ JWT token หลัง login                   |
	 | `room_id`         | `101`                      | ตัวอย่างหมายเลขห้องที่ใช้ทดสอบ                |

5. กด **Save**
6. มุมขวาบนของ Postman เลือก Environment เป็น `Dormitory-Microservices-Local`

---

## 3. สร้าง Collection สำหรับ API ทั้งหมด

แนะนำให้สร้าง Collection เดียวชื่อ `Dormitory Backend (via Gateway)` แล้วแยก folder ตาม service:

- `User`
- `Rooms`
- `Meters`
- `Bills`

ขั้นตอน:

1. กด **New → Collection** ตั้งชื่อ `Dormitory Backend (via Gateway)`
2. คลิกขวาที่ Collection → **Add Folder** สร้างตามรายการด้านบน
3. ทุก request ใน collection นี้ ให้ใช้ URL เริ่มต้นด้วย `{{base_url}}`

---

## 4. ตั้งค่า Login Request (Admin) ให้เก็บ Token อัตโนมัติ

### 4.1 สร้าง Request: Admin Login

1. ใน folder `User` → กด **Add Request** ตั้งชื่อ `Admin Login`
2. Method: `POST`
3. URL:

	 ```text
	 {{base_url}}/api/users/login
	 ```

4. แท็บ **Body** → เลือก `raw` + `JSON` ใส่ค่า:

	 ```json
	 {
		 "username": "{{admin_username}}",
		 "password": "{{admin_password}}"
	 }
	 ```

5. แท็บ **Tests** ใส่ script เพื่อดึง token เก็บลง Environment:

	 ```javascript
	 const res = pm.response.json();
	 if (res.token) {
		 pm.environment.set("token", res.token);
		 console.log("Saved token to environment");
	 }
	 ```

6. กด **Save**

### 4.2 ตั้งค่า Authorization สำหรับ Request อื่น

ทุก request ที่ต้องใช้ JWT ให้ตั้งค่าแบบเดียวกัน:

1. ไปที่ระดับ **Collection** (`Dormitory Backend (via Gateway)`) → แท็บ **Authorization**
2. Type: `Bearer Token`
3. ช่อง **Token** ใส่:

	 ```text
	 {{token}}
	 ```

4. กด **Save**

ตอนนี้ทุก request ใน collection จะใช้ header:

```http
Authorization: Bearer {{token}}
```

โดยอัตโนมัติ (ไม่ต้องตั้งทีละ request)

---

## 5. Flow การ Demo หลัก (Step-by-step)

ด้านล่างเป็นลำดับการยิง API ที่แนะนำเวลา Demo ให้เห็นภาพรวมของระบบ

### 5.1 เริ่มจาก Login เป็น Admin

1. เลือก request `Admin Login`
2. กด **Send**
3. ตรวจดู response ว่ามี field `token`
4. ดูที่ Environment → ตัวแปร `token` ควรถูกเซ็ตอัตโนมัติ

> จากนี้ไป ทุก request ที่อยู่ใน collection จะส่ง JWT นี้ไปที่ Gateway โดยอัตโนมัติ

---

### 5.2 ดูรายชื่อผู้ใช้ (User Service)

สร้าง request ใน folder `User`:

- ชื่อ: `Get All Users`
- Method: `GET`
- URL:

	```text
	{{base_url}}/api/users/users
	```

กด Send:

- จะเห็นรายการ user: `admin`, `tenant1` พร้อม role

จุดที่อธิบายตอน demo:

- Gateway ตรวจ JWT ก่อน แล้ว forward header `X-User-Name`, `X-User-Role` ไปที่ user-service
- user-service ใช้ middleware JWT (เฉพาะ /me, /users) เพื่ออ่านข้อมูลจาก token ตรงๆ

---

### 5.3 ดูข้อมูลห้อง (Room Service)

#### 5.3.1 Get All Rooms

สร้าง request ใน folder `Rooms`:

- ชื่อ: `Get All Rooms`
- Method: `GET`
- URL:

	```text
	{{base_url}}/api/rooms/
	```

ผลลัพธ์:

- เห็นห้องทั้งหมด เช่น 101, 102, 201 พร้อม status และ tenant_name

#### 5.3.2 Get Room Detail

- ชื่อ: `Get Room Detail`
- Method: `GET`
- URL:

	```text
	{{base_url}}/api/rooms/{{room_id}}
	```

ผลลัพธ์:

- เห็นข้อมูลห้องเดียวตาม `room_id`

จุดที่อธิบายตอน demo:

- gateway ตรวจ JWT แล้วส่ง header ต่อไป
- room-service ใช้ `RequireUser` เช็คว่าต้อง login ถึงจะเข้า endpoint เหล่านี้ได้

---

### 5.4 บันทึกมิเตอร์น้ำ/ไฟ (Meter Service)

#### 5.4.1 Record Water (Admin)

- ชื่อ: `Record Water`
- Folder: `Meters`
- Method: `POST`
- URL:

	```text
	{{base_url}}/api/meters/water
	```

- Body → raw JSON:

	```json
	{
		"room_id": "{{room_id}}",
		"unit": 15.5
	}
	```

ผลลัพธ์:

- ได้ object ของ WaterMeter กลับมา พร้อม field `month`
- ใน log ของ meter-service จะเห็นข้อความ
	- `Meter Service: create water meter room_id=101 unit=15.50 by user=admin`
	- `Published event to meter.water.created`

#### 5.4.2 Record Electric (Admin)

- ชื่อ: `Record Electric`
- Method: `POST`
- URL:

	```text
	{{base_url}}/api/meters/electric
	```

- Body:

	```json
	{
		"room_id": "{{room_id}}",
		"unit": 120.0
	}
	```

ผลลัพธ์คล้ายกับน้ำ และมี log publish ไป `meter.electric.created`

#### 5.4.3 ดูค่ามิเตอร์ล่าสุด

- `GET {{base_url}}/api/meters/water/{{room_id}}`
- `GET {{base_url}}/api/meters/electric/{{room_id}}`

ใช้แสดงตอน demo ว่าข้อมูลมิเตอร์ถูกบันทึกและดึงมาใช้คำนวณบิลได้

---

### 5.5 สร้างและดูบิลค่าเช่า (Bill Service)

#### 5.5.1 Create Bill (Admin)

- ชื่อ: `Create Bill`
- Folder: `Bills`
- Method: `POST`
- URL:

	```text
	{{base_url}}/api/bills/{{room_id}}
	```

- Body (จะส่งหรือไม่ส่งก็ได้ เพราะ server คำนวณเอง ถ้าต้องการอัปเดตสถานะส่งเพิ่มได้):

	```json
	{
		"status": "Unpaid"
	}
	```

ผลลัพธ์:

- ได้ object Bill กลับมา มี `rent_price`, `water_price`, `electric_price`, `total`, `status`
- ดู log ของ bill-service:
	- `Bill Service: create bill for room_id=101`
	- `Bill Service: call room-service for room_id=101`
	- `Bill Service: call meter-service for latest water room_id=101`
	- `Bill Service: call meter-service for latest electric room_id=101`
	- `Bill Service: calculated bill for room_id=101 ...`

#### 5.5.2 Get Latest Bill (Tenant/Admin)

- ชื่อ: `Get Latest Bill`
- Method: `GET`
- URL:

	```text
	{{base_url}}/api/bills/{{room_id}}
	```

**กรณี Login เป็น admin:** เห็นได้ทุกห้อง

**กรณี Login เป็น tenant1:**

1. สลับไปใช้ request `Tenant Login` (สร้างเพิ่มแบบเดียวกับ Admin แต่ใช้ `{{tenant_username}}`, `{{tenant_password}}`)
2. ยิง `Tenant Login` → token จะถูกเซ็ตใหม่เป็นของ tenant1
3. ยิง `Get Latest Bill` อีกครั้ง:
	 - ถ้า `room_id` เป็นห้องของ tenant1 (เช่น 101) → เห็นบิลได้
	 - ถ้า `room_id` เป็นห้องอื่น → ได้ 403 "ไม่มีสิทธิ์ดูบิลของห้องนี้"

#### 5.5.3 Get All Bills (Admin)

- ชื่อ: `Get All Bills`
- Method: `GET`
- URL:

	```text
	{{base_url}}/api/bills/
	```

ใช้โชว์ตอน demo ว่า admin เห็นภาพรวมทุกบิลในระบบได้

#### 5.5.4 Update Bill Status (Admin)

- ชื่อ: `Update Bill`
- Method: `PATCH`
- URL:

	```text
	{{base_url}}/api/bills/{{room_id}}
	```

- Body:

	```json
	{
		"status": "Paid"
	}
	```

ผลลัพธ์: สถานะของบิลล่าสุดของห้องนั้นจะถูกเปลี่ยนเป็น `Paid`

---

## 6. Flow แนะนำสำหรับใช้ตอน Demo หน้าเพื่อน/อาจารย์

สามารถเล่าเป็นเรื่องราวแบบนี้:

1. **เริ่มจากภาพรวม**
	 - อธิบายว่าใช้ Microservices 5 ตัว + Gateway + RabbitMQ
	 - ทุก request จาก Postman ยิงเข้า `{{base_url}}` ที่ Gateway ที่เดียว

2. **Login เป็น Admin**
	 - โชว์ request `Admin Login`
	 - ชี้ให้ดู Tests ที่ดึง token แล้วเก็บลง Environment
	 - เปิด Environment ให้ดูว่า `token` ถูกเติมอัตโนมัติ

3. **ดู Users และ Rooms ผ่าน Gateway**
	 - ยิง `Get All Users`
	 - ยิง `Get All Rooms`
	 - อธิบายว่า Gateway ตรวจ JWT แล้วแนบ header `X-User-Name`, `X-User-Role` ให้ service ปลายทาง

4. **บันทึกมิเตอร์น้ำ/ไฟ + RabbitMQ**
	 - ยิง `Record Water` + `Record Electric`
	 - เปิด log meter-service ให้ดูข้อความ `Published event to meter.water.created` เป็นต้น
	 - อธิบายว่า event ถูกส่งไปที่ RabbitMQ และ bill-service subscribe อยู่

5. **สร้างบิล + call ไปหา service อื่น**
	 - ยิง `Create Bill`
	 - เปิด log bill-service ให้ดูว่ามัน call ไป `room-service` และ `meter-service` ตามลำดับ

6. **สลับเป็น Tenant แล้วดูสิทธิ์**
	 - ยิง `Tenant Login` เพื่อเปลี่ยน token เป็นของ tenant1
	 - ยิง `Get Latest Bill` ของห้องตัวเอง → ผ่าน
	 - ลองเปลี่ยน `room_id` เป็นห้องอื่น → ได้ 403

7. (ถ้ามีเวลา) **Demo Circuit Breaker ที่ Gateway**
	 - หยุด container ของ service ใด service หนึ่ง (เช่น room-service)
	 - ใช้ script ยิงไป `GET {{base_url}}/api/rooms/` รัว ๆ
	 - ชี้ให้ดู log ของ Gateway ที่ขึ้นว่า CIRCUIT OPEN และ HTTP 503

---

## 7. สรุป

แค่เตรียม Environment ให้ดี (base_url, token, room_id ฯลฯ) และจัดลำดับ request ตาม flow ด้านบน ก็สามารถ Demo ระบบได้ครบทั้ง

- การ login + สิทธิ์แยก Admin / Tenant
- การเรียก service ผ่าน Gateway เพียงจุดเดียว
- การเชื่อม Room, Meter, Bill เข้าด้วยกัน
- การใช้ RabbitMQ สำหรับ event ของมิเตอร์

ถ้าอยากให้ช่วยทำ Postman Collection (ไฟล์ .json) สำหรับ import โดยตรง สามารถบอกชื่อที่อยากใช้ เดี๋ยวผมเขียนโครงให้ copy-paste ไปสร้าง collection ได้เลย

