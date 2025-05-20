## emission
# Requirements:
### 1. Functional Requirements
✅ Retrieve stored emission data

✅ Store emission data

✅ Config emission factors

✅ Authenticate client
### 2. Non-functional Requirements
✅ Security

- Secure API with API key authentication.

✅ Performance & Scalability

- Scale API servers and NATS consumers dynamically.

✅ Availability & Reliability
- Containerized deployment Docker.
- Use message persistence in NATS to prevent data loss.

✅ Logging & Monitoring
- Use structured logging (Zap).

# System Design:
![Architecture](co2-emission-system.png)

- The system is composed of three main components:
  - **emission-service**: REST API service that provides emission data.
  - **NATS server**: Message broker that keep messages between the producer and the consumer.
  - **emission-calculator**: NATS consumer that listens to the emission data and stores it in the database.


# Data Model:
![database](img.png)

# API Documentation

## **1. Create a Factory**

### **POST /factories**
**Description:**  
Registers a new factory and generates a unique API key for authentication.

### **Request:**
- **Headers:** None
- **Body (JSON):**
  ```json
  {
    "name": "Factory Name",
    "country": "Country Name" 
  }
  ```

### **Response:**
- **Status:** `201 Created`
- **Body (JSON):**
  ```json
  {
    "api_key": "your-generated-api-key"
  }
  ```

---

## **2. Retrieve Emission Data**

### **GET /emissions**
**Description:**  
Fetches emission data for a factory based on its API key. Supports optional filtering.

### **Request:**
- **Headers:**
  ```plaintext
  Authorization: Bearer {API_KEY}
  ```
- **Query Parameters (Optional):**
  - `sort=` (e.g., `calculated_emission`, `-calculated_emission`) - Sort order by created_at, calculated_emission. 
  - `from=` Start date 
  - `to=` End date

**Example Request:**
```
GET /emissions?sort=-calculated_emission&from=2024-12-09T15:04:05Z&to=2024-13-09T15:04:05Z
```

### **Response:**
- **Status:** `200 OK`
- **Body (JSON):**
  ```json
  {
    "emissions": [
      {
        "timestamp": "2024-02-01T10:30:00Z",
        "electricity_consumption": 100.5,
        "emission_factor": 0.417,
        "calculated_emission": 41.87
      }
    ]
  }
  ```

---

## **3. Submit Electricity Consumption Data**

### **POST /emissions**
**Description:**  
Submits electricity consumption data for emissions calculation.

### **Request:**
- **Headers:**
  ```plaintext
  Authorization: Bearer {API_KEY}
  ```
- **Body (JSON):**
  ```json
  {
    "electricity_consumption": 120.5
  }
  ```

### **Response:**
- **Status:** `200 OK`
- **Body (JSON):**
  ```json
      {
        "timestamp": "2024-02-01T10:30:00Z",
        "electricity_consumption": 100.5,
        "emission_factor": 0.417,
        "calculated_emission": 41.87
      }
  ```

---

## **Notes**

✅ **API Key Authentication:**
- All requests requiring authentication must include the factory’s API key in the `Authorization` header as:
  ```plaintext
  Authorization: Bearer {API_KEY}
  ```

✅ **Error Handling:**
- **`400 Bad Request`** → Invalid request format.
- **`401 Unauthorized`** → Missing or invalid API key.
- **`500 Internal Server Error`** → Unexpected system error.  




