CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    role_name VARCHAR(255) UNIQUE NOT NULL,  -- Название роли, например "admin", "hr", "user"
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );

CREATE TABLE IF NOT EXISTS permissions (
    id SERIAL PRIMARY KEY,
    permission_name VARCHAR(255) UNIQUE NOT NULL,  -- Название права, например "assign_test", "view_report"
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );

-- Таблица для связывания ролей и прав
CREATE TABLE IF NOT EXISTS role_permissions (
                                                role_id INT REFERENCES roles(id),
    permission_id INT REFERENCES permissions(id),
    PRIMARY KEY (role_id, permission_id),
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );

CREATE TABLE IF NOT EXISTS users
(
    id SERIAL PRIMARY KEY,
    role_id INT REFERENCES roles(id),  -- Роль пользователя (например, user, admin, etc.)
    telegram_username VARCHAR(255) UNIQUE NOT NULL,
    telegram_first_name VARCHAR(255),
    real_first_name VARCHAR(255),
    real_second_name VARCHAR(255),
    real_surname VARCHAR(255),
    current_state VARCHAR(50),  -- Текущее состояние пользователя (например, testing, finished)
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );

CREATE TABLE IF NOT EXISTS tests
(
    id SERIAL PRIMARY KEY,
    test_name VARCHAR(255),
    test_type VARCHAR(50),  -- Тип теста
    duration INT,  -- Время выполнения теста в минутах
    question_count INT,  -- Количество вопросов в тесте
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );

CREATE TABLE IF NOT EXISTS questions
(
    id SERIAL PRIMARY KEY,
    test_id INT REFERENCES tests(id),  -- Внешний ключ к тесту
    question_text TEXT NOT NULL,  -- Текст вопроса
    answer_type VARCHAR(50),  -- Тип вопроса (например, multiple, single, open)
    correct_answer TEXT,  -- Правильный ответ (для открытых вопросов)
    test_options JSONB,  -- Варианты ответов (для вопросов с несколькими вариантами)
    UNIQUE (test_id, question_text),  -- Уникальность вопросов в рамках одного теста
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );

CREATE TABLE IF NOT EXISTS user_tests
(
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),  -- Внешний ключ к пользователю, может быть NULL для отложенных назначений
    test_id INT REFERENCES tests(id),  -- Внешний ключ к тесту
    assigned_by INT REFERENCES users(id),  -- Внешний ключ к пользователю, который назначил тест
    pending_username VARCHAR(255),  -- Username для отложенных назначений
    current_question_index INT,  -- Индекс текущего вопроса, на котором находится пользователь
    correct_answers_count INT,  -- Количество правильных ответов
    message_id INT,  -- ID сообщения с таймером
    timer_deadline TIMESTAMP,  -- Дедлайн таймера
    start_time TIMESTAMP,  -- Время начала теста
    end_time TIMESTAMP,  -- Время завершения теста
    status VARCHAR(50),  -- Статус теста для пользователя (например, pending, assigned, in_progress, completed)
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_user_or_pending CHECK (
(user_id IS NOT NULL AND pending_username IS NULL) OR
(user_id IS NULL AND pending_username IS NOT NULL)
    )
    );

CREATE TABLE IF NOT EXISTS answers
(
    id SERIAL PRIMARY KEY,
    user_test_id INT REFERENCES user_tests(id),  -- Внешний ключ к записи теста пользователя
    question_id INT REFERENCES questions(id),  -- Внешний ключ к вопросу
    user_answer TEXT,  -- Ответ пользователя
    is_correct BOOLEAN,  -- Является ли ответ правильным
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );

CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    message_key VARCHAR(255) UNIQUE NOT NULL,  -- Уникальный ключ для сообщения (например, "welcome_message")
    message_text TEXT NOT NULL,  -- Текст самого сообщения
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );
