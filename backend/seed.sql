-- Seed data for TaskFlow
-- Password for test@example.com is: password123
-- bcrypt hash of 'password123' with cost 12

INSERT INTO users (id, name, email, password) VALUES (
    'a0000000-0000-0000-0000-000000000001',
    'Test User',
    'test@example.com',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBdXIG/QjPCrAS'
) ON CONFLICT (email) DO NOTHING;

INSERT INTO projects (id, name, description, owner_id) VALUES (
    'b0000000-0000-0000-0000-000000000001',
    'Demo Project',
    'A sample project to get you started with TaskFlow',
    'a0000000-0000-0000-0000-000000000001'
) ON CONFLICT DO NOTHING;

INSERT INTO tasks (title, description, status, priority, project_id, assignee_id) VALUES
    (
        'Design the homepage',
        'Create wireframes and mockups for the main landing page',
        'todo',
        'high',
        'b0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        'Set up CI/CD pipeline',
        'Configure GitHub Actions for automated testing and deployment',
        'in_progress',
        'medium',
        'b0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    ),
    (
        'Write API documentation',
        'Document all REST endpoints with request/response examples',
        'done',
        'low',
        'b0000000-0000-0000-0000-000000000001',
        'a0000000-0000-0000-0000-000000000001'
    );
