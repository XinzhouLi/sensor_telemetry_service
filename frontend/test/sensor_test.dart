import 'package:flutter_test/flutter_test.dart';
import 'package:sensor_dashboard/sensor.dart';

void main() {
  test('parses a sensor with its latest reading', () {
    final sensor = Sensor.fromJson({
      'id': 'nox-analyzer-1',
      'name': 'NOx Analyzer 1',
      'unit': 'ppm',
      'health': 'ok',
      'latest_reading': {
        'recorded_at': '2026-07-20T08:03:00-06:00',
        'value': 41.2,
        'status': 'valid',
      },
    });

    expect(sensor.id, 'nox-analyzer-1');
    expect(sensor.latestReading?.value, 41.2);
    expect(sensor.latestReading?.status, 'valid');
    expect(
      sensor.latestReading?.recordedAt.toUtc(),
      DateTime.parse('2026-07-20T14:03:00Z'),
    );
  });

  test('parses a sensor that has never reported', () {
    final sensor = Sensor.fromJson({
      'id': 'stack-temp-1',
      'name': 'Stack Temperature 1',
      'unit': '°C',
      'health': 'never_reported',
      'latest_reading': null,
    });

    expect(sensor.latestReading, isNull);
    expect(sensor.health, 'never_reported');
  });

  test('parses valid and out-of-range readings', () {
    final readings = [
      Reading.fromJson({
        'recorded_at': '2026-07-20T08:03:00-06:00',
        'value': 41.2,
        'status': 'valid',
      }),
      Reading.fromJson({
        'recorded_at': '2026-07-20T14:04:00Z',
        'value': 512,
        'status': 'out_of_range',
      }),
    ];

    expect(
      readings[0].recordedAt.toUtc().toIso8601String(),
      '2026-07-20T14:03:00.000Z',
    );
    expect(readings[0].status, 'valid');
    expect(readings[1].value, 512);
    expect(readings[1].status, 'out_of_range');
  });
}
