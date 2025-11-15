-- Migration script to create all tables with English names and localization support
-- This script uses PostgreSQL syntax and is idempotent

-- Create enum type for auditorium type (English values only)
DO $$ 
BEGIN
    CREATE TYPE auditorium_type_enum AS ENUM ('coworking', 'classroom', 'lecture_hall');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Create City table
-- Localized fields: name_ru, name_en (no base 'name' column)
CREATE TABLE IF NOT EXISTS City (
    id SERIAL PRIMARY KEY,
    name_ru VARCHAR(255) NOT NULL,
    name_en VARCHAR(255) NOT NULL
);

-- Create Building table
-- Localized fields: address_ru, address_en (no base 'address' column)
CREATE TABLE IF NOT EXISTS Building (
    id SERIAL PRIMARY KEY,
    city_id INTEGER NOT NULL,
    address_ru VARCHAR(255) NOT NULL,
    address_en VARCHAR(255) NOT NULL,
    floor_count INTEGER NOT NULL,
    CONSTRAINT fk_building_city FOREIGN KEY (city_id) REFERENCES City(id) ON DELETE CASCADE
);

-- Create Auditorium table
-- Type field: enum 'type' column plus localized type_ru and type_en columns
CREATE TABLE IF NOT EXISTS Auditorium (
    id SERIAL PRIMARY KEY,
    building_id INTEGER NOT NULL,
    floor_number INTEGER NOT NULL,
    capacity INTEGER NOT NULL,
    auditorium_number VARCHAR(50) NOT NULL,
    type auditorium_type_enum NOT NULL,
    type_ru VARCHAR(50) NOT NULL,
    image_url VARCHAR(500),
    CONSTRAINT fk_auditorium_building FOREIGN KEY (building_id) REFERENCES Building(id) ON DELETE CASCADE
);

-- Create Occupancy table
CREATE TABLE IF NOT EXISTS Occupancy (
    id SERIAL PRIMARY KEY,
    auditorium_id INTEGER NOT NULL,
    person_count INTEGER NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    CONSTRAINT fk_occupancy_auditorium FOREIGN KEY (auditorium_id) REFERENCES Auditorium(id) ON DELETE CASCADE
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_building_city_id ON Building(city_id);
CREATE INDEX IF NOT EXISTS idx_auditorium_building_id ON Auditorium(building_id);
CREATE INDEX IF NOT EXISTS idx_auditorium_number ON Auditorium(auditorium_number);
CREATE INDEX IF NOT EXISTS idx_occupancy_auditorium_id ON Occupancy(auditorium_id);
CREATE INDEX IF NOT EXISTS idx_occupancy_timestamp ON Occupancy(timestamp);

-- Optional: Insert sample data (uncomment to use)

-- Sample Cities
INSERT INTO City (name_ru, name_en) VALUES 
    ('Москва', 'Moscow'),
    ('Санкт-Петербург', 'Saint Petersburg'),
    ('Новосибирск', 'Novosibirsk')
ON CONFLICT DO NOTHING;

-- Sample Buildings
INSERT INTO Building (city_id, address_ru, address_en, floor_count) VALUES 
    (1, 'ул. Ленина, д. 1', 'Lenina St, 1', 5),
    (1, 'пр. Мира, д. 10', 'Mira Ave, 10', 3),
    (2, 'Невский проспект, д. 20', 'Nevsky Prospect, 20', 4),
    (3, 'Красный проспект, д. 50', 'Krasny Prospect, 50', 6)
ON CONFLICT DO NOTHING;

-- Sample Auditoriums
INSERT INTO Auditorium (
    building_id, 
    floor_number, 
    capacity, 
    auditorium_number, 
    type, 
    type_ru, 
    image_url
) VALUES 
    (1, 1, 30, '101', 'classroom', 'учебная',  'https://example.com/images/101.jpg'),
    (1, 2, 50, '201', 'lecture_hall', 'лекционная',  'https://example.com/images/201.jpg'),
    (1, 3, 20, '301', 'coworking', 'коворкинг',  'https://example.com/images/301.jpg'),
    (2, 1, 40, '101', 'classroom', 'учебная',  'https://example.com/images/102.jpg'),
    (2, 2, 60, '201', 'lecture_hall', 'лекционная',  'https://example.com/images/202.jpg'),
    (3, 2, 100, '201', 'lecture_hall', 'лекционная',  'https://example.com/images/203.jpg'),
    (4, 1, 25, '101', 'classroom', 'учебная', 'https://example.com/images/104.jpg')
ON CONFLICT DO NOTHING;

-- Sample Occupancy records
INSERT INTO Occupancy (auditorium_id, person_count, timestamp) VALUES 
    (1, 25, '2024-01-15 10:30:00+00'),
    (1, 30, '2024-01-15 12:00:00+00'),
    (2, 45, '2024-01-15 14:00:00+00'),
    (3, 15, '2024-01-15 16:00:00+00')
ON CONFLICT DO NOTHING;


