CREATE TYPE reading_status AS ENUM ('valid', 'out_of_range');

CREATE TABLE sensors (
    id text PRIMARY KEY,
    name text NOT NULL,
    unit text NOT NULL,
    valid_min double precision NOT NULL,
    valid_max double precision NOT NULL,
    CONSTRAINT sensors_valid_range CHECK (valid_min <= valid_max)
);

CREATE TABLE readings (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    sensor_id text NOT NULL,
    recorded_at timestamptz NOT NULL,
    value double precision NOT NULL,
    status reading_status NOT NULL,
    CONSTRAINT readings_sensor_foreign_key
        FOREIGN KEY (sensor_id) REFERENCES sensors(id) ON DELETE RESTRICT,
    CONSTRAINT readings_sensor_recorded_at_unique
        UNIQUE (sensor_id, recorded_at)
);
