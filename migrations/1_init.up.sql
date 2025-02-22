CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    role_name VARCHAR(255) UNIQUE NOT NULL,  -- Название роли, например "admin", "hr", "user"
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );

CREATE TABLE IF NOT EXISTS permissions (
    id SERIAL PRIMARY KEY,
    permission_name VARCHAR(255) UNIQUE NOT NULL,  -- Название права, например "assign_test", "view_report"
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );

-- Таблица для связывания ролей и прав
CREATE TABLE IF NOT EXISTS role_permissions (
                                                role_id INT REFERENCES roles(id),
    permission_id INT REFERENCES permissions(id),
    PRIMARY KEY (role_id, permission_id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );

CREATE TABLE IF NOT EXISTS users
(
    id SERIAL PRIMARY KEY,
    role_id INT REFERENCES roles(id),
    telegram_username VARCHAR(255) UNIQUE NOT NULL,
    telegram_id BIGINT UNIQUE,
    telegram_first_name VARCHAR(255),
    real_first_name VARCHAR(255),
    real_second_name VARCHAR(255),
    real_surname VARCHAR(255),
    current_state VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tests
(
    id SERIAL PRIMARY KEY,
    test_name VARCHAR(255),
    test_type VARCHAR(50),
    duration INT,
    question_count INT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );

CREATE TABLE IF NOT EXISTS questions
(
    id SERIAL PRIMARY KEY,
    test_id INT REFERENCES tests(id),
    question_text TEXT NOT NULL,
    answer_type VARCHAR(50),
    correct_answer TEXT,
    test_options JSONB,
    UNIQUE (test_id, question_text),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );

CREATE TABLE IF NOT EXISTS user_tests
(
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    test_id INT REFERENCES tests(id),
    assigned_by INT REFERENCES users(id),
    pending_username VARCHAR(255),
    current_question_index INT,
    correct_answers_count INT,
    message_id INT,
    timer_deadline TIMESTAMP,
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    status VARCHAR(50),
    selected_question_ids INTEGER[] DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_user_or_pending CHECK (
(user_id IS NOT NULL AND pending_username IS NULL) OR
(user_id IS NULL AND pending_username IS NOT NULL)
    )
    );

CREATE TABLE IF NOT EXISTS answers
(
    id SERIAL PRIMARY KEY,
    user_test_id INT REFERENCES user_tests(id),
    question_id INT REFERENCES questions(id),
    user_answer TEXT,
    is_correct BOOLEAN,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    message_key VARCHAR(255) UNIQUE NOT NULL,
    message_text TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS test_links (
    id SERIAL PRIMARY KEY,
    test_id INT REFERENCES tests(id),
    token VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);