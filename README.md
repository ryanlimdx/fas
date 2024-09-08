# Financial Assistance Scheme Management System

This is a backend application that will be part of a system to allow administrators to manage financial assistance schemes and applications for schemes.

The API documentation is located [here](https://documenter.getpostman.com/view/38191594/2sAXjRWVTM#fa66d61e-4de5-4ec6-a4b8-dbcbc8727466).

## Prerequisites

Before you begin, ensure you have met the following requirements:

- **Go** (version 1.23.0 or later): [Download Go](https://golang.org/dl/)
- **MySQL** (version 5.7 or later): [Download MySQL](https://dev.mysql.com/downloads/mysql/)

Optionally, the following tool is useful to facilitate testing of API endpoints:


- **Postman** (latest version): [Download Postman](https://www.postman.com/downloads/)

## Setting Up Your Local Environment

### Step 1: Clone the Repository

Clone this repository to your local machine using:

```bash
git clone https://github.com/ryanlimdx/fas.git
cd fas
```

Note: This is only if you did not fork this repository. Otherwise, clone this repository to your local machine using:

```bash
git clone https://github.com/yourusername/yourprojectname.git
cd yourprojectname
```

### Step 2: Install Dependencies

Install all necessary dependencies to run the project:

```bash
go mod tidy
```

This command installs all the necessary Go modules as specified in `go.mod`.

### Step 3: Environment Configuration

Create a `.env` file in the `cmd/server` directory of the project to store environment variables (the database connection settings). If you are unsure, feel free to follow the `.env.example` file located in the `cmd/server` directory.

```makefile
DSN=username:password@tcp(hostname:port)/database_name
```

Ensure you replace `username`, `password`, `hostname`, `port`, and `database_name` with your actual MySQL details (should you run the scripts in the next step, `database_name` should be `fas_database`). In case your password contains special characters, do remember to include the escape sequence.

### Step 4: Database Setup

Run the SQL scripts to create the necessary database and tables. You can find the SQL scripts in the init.sql file at `scripts/database`, or you can set them up manually:

```sql
-- Run this SQL in your MySQL client
CREATE DATABASE IF NOT EXISTS fas_database;
USE fas_database;
```

### Step 5: Run the Application

Execute the following command in `cmd/server` to start the server:

```bash
go run main.go
```

This command will also create the necessary tables on start-up. Should a Firewall prompt pop up, do allow the connection. It is to enable the MySQL connection.

## Testing the API Endpoints

To test the API endpoints, you can use **Postman**. Start Postman and import the [collection](https://documenter.getpostman.com/view/38191594/2sAXjRWVTM#fa66d61e-4de5-4ec6-a4b8-dbcbc8727466) or manually create requests to the following URL:


```bash
http://localhost:8080/api/path
```

Replace `/path` with actual endpoints such as `/applicants`, `/applications` or `/schemes` to interact with the API.

The documentation for the API endpoints are located [here](https://documenter.getpostman.com/view/38191594/2sAXjRWVTM#fa66d61e-4de5-4ec6-a4b8-dbcbc8727466).

## Appendix: Database Design Considerations
The database schema for the Financial Assistance Scheme Management System was designed with several considerations:

1. Normalization, Referential Integrity and Data Consistency

    The schema is normalized to avoid data redundancy and ensure referential integrity. Foreign keys are used to maintain relationships between entities like applicants, schemes, criteria, benefits, and applications. These foreign keys enforce referential integrity and ensure that cascading deletes maintain the integrity of data across related tables. Cascading deletes also automatically remove related records when a parent entity is deleted, thus preventing orphaned records and ensuring data consistency. 

2. Primary and Unique Constraints

    Each table has a primary key for unique identification, and unique constraints are applied where necessary to avoid duplicate entries. For instance, in the applications table, the combination of applicant_id and scheme_id is unique, preventing duplicate applications for the same scheme by the same applicant.

3. Handling Complex Relationships

    Many-to-many relationships, such as those between schemes and criteria, are managed using join tables like scheme_criteria and scheme_benefits. This approach allows for flexibility in linking schemes to different criteria and benefits, while reducing duplicate entries for benefits and criterias.

4. Scalability, Flexibility and Future-Proofing

    The use of UUIDs (VARCHAR(36)) for primary keys ensures scalability and global uniqueness across distributed systems. The schema is designed to allow for easy extension, supporting future requirements like new scheme benefits or eligibility criteria; The criteria table allows for flexible eligibility conditions by storing a combination of criteria_level, criteria_type, and status. This makes it easier to extend the system in the future without modifying the schema, ensuring that schemes can evolve without the need for structural changes to the database.
