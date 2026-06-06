# MOH SSO Dashboard

A centralized **Single Sign-On (SSO)** platform for Ministry of Health (MOH) applications.  
Built with **Keycloak**, **Go (Gin)**, **React**, and **Carbon Design System**.

---

## Table of Contents

- [MOH SSO Dashboard](#moh-sso-dashboard)
  - [Table of Contents](#table-of-contents)
  - [1. Introduction](#1-introduction)
  - [2. System Architecture Overview](#2-system-architecture-overview)
    - [2.1 Frontend (React + Carbon)](#21-frontend-react--carbon)
    - [2.2 Backend (Go + Gin)](#22-backend-go--gin)
    - [2.3 Keycloak (Identity Provider)](#23-keycloak-identity-provider)
  - [3. Features](#3-features)
    - [3.1 Authentication \& Authorization](#31-authentication--authorization)
    - [3.2 Client Management](#32-client-management)
    - [3.3 User Management](#33-user-management)
  - [4. API Reference](#4-api-reference)
    - [Authentication API](#authentication-api)
    - [Client API](#client-api)
    - [User API](#user-api)
  - [5. Development Setup](#5-development-setup)
    - [Step 1 — Clone Repository](#step-1--clone-repository)
    - [Step 2 — Backend Setup](#step-2--backend-setup)
    - [Step 3 — Frontend Setup](#step-3--frontend-setup)
  - [6. Docker Deployment](#6-docker-deployment)
    - [Step 1 — Start Development **Stack**](#step-1--start-development-stack)
    - [Step 2 — **Production** Build](#step-2--production-build)
  - [7. Environment Variables](#7-environment-variables)
    - [Backend `.env`](#backend-env)
    - [Frontend `.env`](#frontend-env)
  - [8. Database Schema Overview](#8-database-schema-overview)
  - [9. Keycloak Realm Requirements](#9-keycloak-realm-requirements)
    - [Realm](#realm)
    - [Roles](#roles)
    - [Dashboard Client](#dashboard-client)
  - [10. Future Enhancements](#10-future-enhancements)
  - [11. License](#11-license)

---

## 1. Introduction

The **MOH SSO Dashboard** unifies internal Ministry of Health systems under a centralized authentication and authorization platform.  
Users can access multiple applications using a single login through **Keycloak**.

This repository contains the full implementation of the frontend, backend, and deployment configuration.

---

## 2. System Architecture Overview

The system consists of three major components:

### 2.1 Frontend (React + Carbon)

- React 19 + TypeScript  
- Carbon Design System  
- AuthContext for token management  
- Dynamic application tiles based on user role  
- Fetches user profile using JWT from backend  

---

### 2.2 Backend (Go + Gin)

Core responsibilities:
- Authentication (`/auth/me`, `/auth/logout`)  
- Client management (`/clients`)  
- User management (`/users`)  
- SQLC for type-safe database queries  
- Keycloak Admin API integration  

Technologies:
- Gin Framework  
- PostgreSQL  
- SQLC  
- Zerolog  

---

### 2.3 Keycloak (Identity Provider)

Handles:
- Authentication (OIDC)  
- User roles  
- Service accounts  
- Client configurations  
- Token generation & validation  

---

## 3. Features

### 3.1 Authentication & Authorization
- Authorization Code Flow via Keycloak  
- Token validation middleware  
- Two main roles: `admin`, `user`  
- Admins manage clients and users  
- Users only see assigned applications  

### 3.2 Client Management
Admins can:
- Create/update/delete clients  
- Sync with Keycloak + local DB  
- Configure name, icon, base URL, visibility  
- Handle public/private client modes  

Client Model:
```
ID
ClientID
Name
Description
BaseURL
Icon
PublicClient
Enabled
Attributes
```

### 3.3 User Management
Admins can:
- Create users (Keycloak + DB)  
- Assign roles  
- Enable/disable accounts  
- Edit profile data  
- Delete accounts  

---

## 4. API Reference

### Authentication API
```
GET  /api/v1/auth/me       -> Get authenticated user profile
POST /api/v1/auth/logout   -> Logout
```

### Client API
```
GET    /api/v1/clients       -> List all clients
POST   /api/v1/clients       -> Create client
PUT    /api/v1/clients/:id   -> Update client
DELETE /api/v1/clients/:id   -> Delete client
```

### User API
```
POST   /api/v1/users       -> Create user
GET    /api/v1/users       -> List users
GET    /api/v1/users/:id   -> Get user
DELETE /api/v1/users/:id   -> Delete user
```

---

## 5. Development Setup

### Step 1 — Clone Repository
```bash
git clone https://github.com/moh-sso-dashboard.git
cd sso-dashboard
```

### Step 2 — Backend Setup
Install dependencies:
```bash
go mod tidy
```

Run server:
```bash
go run ./cmd/server
```

### Step 3 — Frontend Setup
```bash
npm install
npm run dev
```

---

## 6. Docker Deployment

### Step 1 — Start Development **Stack**
```bash
sudo docker compose -f docker-compose.dev.yml up -d --build
```

Starts:
- Keycloak  
- PostgreSQL  
- Backend  
- Frontend  

### Step 2 — **Production** Build
```bash
docker build -t moh-sso-backend ./backend
docker build -t moh-sso-frontend ./frontend
```

---

## 7. Environment Variables

### Backend `.env`
```env
DB_SOURCE=postgresql://postgres:postgres@db:5432/sso?sslmode=disable
KEYCLOAK_BASE_URL=http://keycloak:8080
KEYCLOAK_REALM=moh-realm
KEYCLOAK_CLIENT_ID=dashboard
KEYCLOAK_CLIENT_SECRET=changeme
SERVER_PORT=9000
```

### Frontend `.env`
```env
VITE_API_BASE=http://localhost:9000/api/v1
VITE_KEYCLOAK_URL=http://localhost:8081
VITE_KEYCLOAK_REALM=moh-realm
VITE_KEYCLOAK_CLIENT_ID=dashboard
```

---

## 8. Database Schema Overview

Main tables:
- `user`  
- `client`  
- `client_roles`  
- `activity_logs`  
- `audit_trails`  

SQLC output directory:
```
internal/db/sqlc
```

---

## 9. Keycloak Realm Requirements

Your realm must include:

### Realm
```
moh-realm
```

### Roles
```
admin
user
```

### Dashboard Client
```
clientId: dashboard
serviceAccountsEnabled: true
redirectUris: ["http://localhost:3000/*"]
webOrigins: ["*"]
```

---

## 10. Future Enhancements
- Multi-tenant support  
- Analytics dashboard  
- Audit log UI  
- Invite-based onboarding  
- Role-based UI customization  

---

## 11. License
```
MIT License
```
