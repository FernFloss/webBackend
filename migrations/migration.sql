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

-- Insert sample data for city
INSERT INTO city (id, name_ru, name_en) VALUES
(1, 'Москва', 'Moscow'),
(2, 'Санкт-Петербург', 'Saint-Petersberg'),
(3, 'Пермь', 'Perm'),
(4, 'Нижний Новгород', 'Nizhny Novgorod')
ON CONFLICT (id) DO NOTHING;

-- Insert sample data for building
INSERT INTO building (id, city_id, address_ru, address_en, floor_count) VALUES
(1, 1, 'Ул. Таллинская, 34', '34 Tallinskaya Street', 7),
(2, 1, 'Покровский бульвар, 11', '11 Pokrovskiy Bulvar', 7),
(3, 1, 'Ул. Мясницкая, 20', '20 Myasnitskaya Street', 5),
(4, 1, 'Кривоколенный Переулок, 3', ' 3 Krivokolenny Pereulok', 4),
(5, 1, 'Армянский переулок', '4 Armyanskiy pereulok, bldg. 2', 4),
(6, 1, 'Ул. Старая Басманная, 21/4, к. 1', ' 21/4 Staraya Basmannaya, bldg. 1', 5),
(7, 1, 'Ул. Старая Басманная, 21/4, к. 5', ' 21/4 Staraya Basmannaya, bldg. 5', 8),
(8, 1, 'Ул. Шаболовка, 26, к. 2', '26 Shabolovka Street, bldg. 2)', 3),
(9, 1, 'Ул. Шаболовка, 26, к. 3', '26 Shabolovka Street, bldg. 3', 4),
(10, 1, 'Ул. Шаболовка, 26/11, к. 4', '26/11 Shabolovka Street, bldg. 4', 3),
(11, 1, 'Ул. Шаболовка, 26/11, к. 9', '26/11 Shabolovka Street, bldg. 9', 3),
(12, 2, 'Васильевский остров, 25-я линия, 6, к. 1', '6, 25th Liniya, Vasilievsky Ostrov, bldg. 1', 4),
(13, 2, 'Канала Грибоедова наб., 119-121', '119-121 Kanala Griboedova Embankment', 3),
(14, 2, 'Ул. Промышленная, 17', '17 Promyshlennaya Street', 5),
(15, 2, 'Ул. Союза Печатников, 16', '16 Soyuza Pechatnikov Street', 4),
(16, 3, 'Ул. Студенческая, 38, к. 1', '38 Studencheskaya Street, bldg. 1', 4),
(17, 3, 'Гагарина бульвар, 37', '37 Gagarina Bulvar, bldg. 2', 4),
(18, 3, 'Гагарина бульвар, 37а', '37A Gagarina Bulvar, bldg. 3', 4),
(19, 4, 'Ул. Родионова, 13б', '13B Rodionova Street', 4),
(20, 4, 'Ул. Львовская, 1в', '1В Lvovskaya Street', 4),
(21, 4, 'Ул. Большая Печерская, 25/12', '25/12 Bolshaya Pecherskaya Street', 4)
ON CONFLICT (id) DO NOTHING;

-- Insert sample data for auditorium
INSERT INTO auditorium (id, building_id, floor_number, capacity, auditorium_number, type, type_ru, image_url) VALUES
(1, 1, 5, 150, '506', 'lecture_hall', 'лекционная', 'https://example.com/images/1.jpg'),
(2, 1, -1, 120, 'Актовый зал', 'coworking', 'коворкинг', 'https://example.com/images/2.jpg'),
(3, 1, 3, 30, '306', 'classroom', 'учебная', 'https://example.com/images/3.jpg'),
(4, 1, 3, 30, '308', 'classroom', 'учебная', 'https://example.com/images/4.jpg'),
ON CONFLICT (id) DO NOTHING;
