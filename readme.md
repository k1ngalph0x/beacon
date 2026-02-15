Beacon

Beacon is a distributed error tracking system inspired by tools like Sentry.
It collects application events, groups them into issues, and triggers alerts based on configurable rules.

The system is designed as a set of independent services communicating via Kafka.

Architecture Overview

High-level flow:

SDK → Ingestion Service → Kafka (beacon-events)
↓
Issue Service
↓
Kafka (issue-updates)
↓
Alert Service

Core Services

Auth Service

User signup / login

JWT authentication

Project creation

API key generation

Ingestion Service

Receives events from SDK

Validates project keys

Publishes events to Kafka

Issue Service

Consumes events from Kafka

Groups events into issues using fingerprinting

Tracks occurrence count

Publishes issue updates

Alert Service

Consumes issue updates

Evaluates alert rules

Triggers alerts when thresholds are met

Tech Stack

Go

Gin (HTTP framework)

Kafka (event streaming)

PostgreSQL

GORM

JWT (authentication)

Docker (for Kafka/Postgres, optional)

Key Design Decisions
Event-Driven Architecture

Services communicate asynchronously via Kafka.
This allows:

Loose coupling between services

Independent scaling

Failure isolation

Clear data flow

Multi-Tenant Project Model

Each user can create multiple projects.

Each project has public and secret API keys.

Events are partitioned by project_id in Kafka for ordering guarantees.

Issue Grouping

Issues are grouped using a SHA-256 fingerprint of:

message + stack_trace

This allows:

Deduplication of repeated errors

Efficient tracking of frequency

Alert threshold evaluation

Separation of Responsibilities

Ingestion does not process issues.

Issue service does not trigger alerts directly.

Alert service reacts to issue updates via Kafka.

This avoids tight coupling and keeps services independently scalable.

Repository Structure
sdk/
services/
auth-service/
ingestion-service/
issue-service/
alert-service/

Environment Configuration

Each service reads configuration from .env.

Example:

DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=postgres
DB_PASSWORD=postgres
DB_NAME=beacon

JwtKey=your_secret_key

KAFKA_BROKERS=localhost:9092

Running the System

1. Start Infrastructure

Start Kafka and PostgreSQL (via Docker or local installation).

Kafka must expose:

localhost:9092

Postgres must be running and accessible.

2. Start Services

Run each service separately:

cd services/auth-service
go run main.go

cd services/ingestion-service
go run main.go

cd services/issue-service
go run main.go

cd services/alert-service
go run main.go

3. SDK Usage

Initialize SDK with project keys and ingestion endpoint.

Events are sent to:

POST /events

on the ingestion service.

APIs
Auth Service

POST /auth/signup

POST /auth/signin

POST /auth/refresh

POST /user/project

Issue Service

GET /projects/:project_id/issues

GET /issues/:id

PATCH /issues/:id/resolve

Highlights

Event-driven microservice architecture

Kafka-based asynchronous communication

Project-based multi-tenancy

Issue fingerprinting for deduplication

Alert rules triggered by issue updates

JWT-secured APIs

Config-driven infrastructure (no hardcoded brokers)

Scope

This project focuses on:

Distributed system design

Event streaming architecture

Service isolation

Multi-tenant design

It does not currently include:

Rate limiting

Web UI

Horizontal autoscaling configuration

Production deployment manifests

Future Improvements (Optional)

Metrics service

Real-time WebSocket notifications

Role-based access control

Alert delivery integrations (email, Slack)

Observability instrumentation

This project is intended as a backend systems design and implementation exercise focused on building a clean, event-driven architecture using Go and Kafka.
