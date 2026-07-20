INSERT INTO sensors (id, name, unit, valid_min, valid_max) VALUES
    ('nox-analyzer-1', 'NOx Analyzer 1', 'ppm', 0, 250),
    ('o2-analyzer-1', 'Oxygen Analyzer 1', '%', 0, 25),
    ('stack-temp-1', 'Stack Temperature 1', '°C', 0, 600)
ON CONFLICT (id) DO NOTHING;

-- Negative IDs are reserved for development fixtures. They keep this seed
-- idempotent without consuming or colliding with the positive identity sequence.
INSERT INTO readings (id, sensor_id, recorded_at, value, status)
OVERRIDING SYSTEM VALUE
VALUES
    (-1, 'nox-analyzer-1', date_trunc('minute', now()) - interval '5 minutes', 41.2, 'valid'),
    (-2, 'nox-analyzer-1', date_trunc('minute', now()) - interval '10 minutes', 512.0, 'out_of_range'),
    (-3, 'nox-analyzer-1', date_trunc('minute', now()) - interval '65 minutes', 40.9, 'valid'),
    (-4, 'o2-analyzer-1', date_trunc('minute', now()) - interval '30 minutes', 20.8, 'valid')
ON CONFLICT (id) DO NOTHING;
