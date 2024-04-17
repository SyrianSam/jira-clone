CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    title TEXT,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    birth_date DATE,
    email VARCHAR(255),
    postal_code VARCHAR(20),
    city VARCHAR(255),
    regulatory_compliance_check BOOLEAN,
    contract_compliance BOOLEAN,
    task_creator VARCHAR(255),
    task_responsible VARCHAR(255),
    assigned_to INTEGER,
    description TEXT,
    comments TEXT,
    state TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
