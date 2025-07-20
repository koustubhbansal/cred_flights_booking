-- Create flights table
CREATE TABLE IF NOT EXISTS flights (
    id SERIAL PRIMARY KEY,
    flight_number VARCHAR(20) NOT NULL,
    source VARCHAR(3) NOT NULL,
    destination VARCHAR(3) NOT NULL,
    departure_time TIMESTAMP NOT NULL,
    arrival_time TIMESTAMP NOT NULL,
    total_seats INTEGER NOT NULL,
    booked_seats INTEGER DEFAULT 0,
    price DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create bookings table
CREATE TABLE IF NOT EXISTS bookings (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    flight_id INTEGER NOT NULL,
    seats INTEGER NOT NULL,
    total_amount DECIMAL(10,2) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    payment_id VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (flight_id) REFERENCES flights(id)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_flights_source_dest_date ON flights(source, destination, departure_time);
CREATE INDEX IF NOT EXISTS idx_flights_source ON flights(source);
CREATE INDEX IF NOT EXISTS idx_bookings_user_id ON bookings(user_id);
CREATE INDEX IF NOT EXISTS idx_bookings_status ON bookings(status);

-- Insert sample flight data
INSERT INTO flights (flight_number, source, destination, departure_time, arrival_time, total_seats, booked_seats, price) VALUES
-- Direct flights from DEL to BOM
('AI101', 'DEL', 'BOM', '2024-02-15 08:00:00', '2024-02-15 10:30:00', 180, 50, 8500.00),
('AI102', 'DEL', 'BOM', '2024-02-15 14:00:00', '2024-02-15 16:30:00', 180, 30, 9200.00),
('AI103', 'DEL', 'BOM', '2024-02-15 20:00:00', '2024-02-15 22:30:00', 180, 80, 7800.00),

-- Direct flights from DEL to BLR
('AI201', 'DEL', 'BLR', '2024-02-15 09:00:00', '2024-02-15 12:00:00', 180, 40, 12000.00),
('AI202', 'DEL', 'BLR', '2024-02-15 15:00:00', '2024-02-15 18:00:00', 180, 60, 13500.00),

-- Direct flights from BOM to BLR
('AI301', 'BOM', 'BLR', '2024-02-15 10:00:00', '2024-02-15 11:30:00', 180, 25, 6500.00),
('AI302', 'BOM', 'BLR', '2024-02-15 16:00:00', '2024-02-15 17:30:00', 180, 45, 7200.00),

-- Connecting flights for multi-stop routes
-- DEL -> BOM -> BLR (for DEL to BLR via BOM)
('AI401', 'DEL', 'HYD', '2024-02-15 07:00:00', '2024-02-15 09:00:00', 180, 20, 9500.00),
('AI402', 'HYD', 'BLR', '2024-02-15 10:30:00', '2024-02-15 11:30:00', 180, 35, 5500.00),

-- BOM -> HYD -> BLR (for BOM to BLR via HYD)
('AI403', 'BOM', 'HYD', '2024-02-15 08:30:00', '2024-02-15 10:00:00', 180, 30, 6800.00),
('AI404', 'HYD', 'BLR', '2024-02-15 11:00:00', '2024-02-15 12:00:00', 180, 25, 5500.00),

-- More connecting flights for complex routes
('AI405', 'DEL', 'CCU', '2024-02-15 06:00:00', '2024-02-15 08:00:00', 180, 15, 11000.00),
('AI406', 'CCU', 'BLR', '2024-02-15 09:30:00', '2024-02-15 12:30:00', 180, 20, 8500.00),

-- Return flights
('AI501', 'BOM', 'DEL', '2024-02-15 11:00:00', '2024-02-15 13:30:00', 180, 40, 8500.00),
('AI502', 'BLR', 'DEL', '2024-02-15 13:00:00', '2024-02-15 16:00:00', 180, 35, 12000.00),
('AI503', 'BLR', 'BOM', '2024-02-15 12:00:00', '2024-02-15 13:30:00', 180, 30, 6500.00); 