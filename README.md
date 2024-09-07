# Project Name

Brief description of what the project does and who it's for.

## Prerequisites

Before you begin, ensure you have met the following requirements:
- **Go** (version 1.23.0 or later): [Download Go](https://golang.org/dl/)
- **MySQL** (version 5.7 or later): [Download MySQL](https://dev.mysql.com/downloads/mysql/)
- Any other dependencies or environmental prerequisites.

## Setting Up Your Local Environment

### Step 1: Clone the Repository

Clone this repository to your local machine using:

```bash
git clone https://github.com/yourusername/yourprojectname.git
cd yourprojectname
```

### Step 2: Install Dependencies
Install all necessary dependencies to run the project:

```bash
go mod tidy
```

This command installs all the necessary Go modules as specified in `go.md`.

### Step 3: Environment Configuration
Create a `.env` file in the `cmd/server` directory of the project to store environment variables (the database connection settings). If you are unsure, feel free to follow the `.env.example` file located in the `cmd/server` directory.

```makefile
DNS=username:password@tcp(hostname:port)/database_name
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

This command will also create the necessary tables on start-up.

## Testing

To test the application, navigate to:

```bash
http://localhost:8080/api/path
```

Replace `/path` with actual endpoints to interact with the API.