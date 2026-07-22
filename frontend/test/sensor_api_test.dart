import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:sensor_dashboard/sensor_api.dart';

void main() {
  test('loads sensors from the overview endpoint', () async {
    final client = MockClient((request) async {
      expect(request.url.path, '/api/sensors');
      return http.Response('''
        [{
          "id": "nox-analyzer-1",
          "name": "NOx Analyzer 1",
          "unit": "ppm",
          "health": "ok",
          "latest_reading": null
        }]
      ''', 200);
    });

    final sensors = await SensorApi(client).listSensors();

    expect(sensors, hasLength(1));
    expect(sensors.single.id, 'nox-analyzer-1');
  });

  test('rejects a non-success response', () async {
    final client = MockClient((_) async => http.Response('unavailable', 503));

    expect(SensorApi(client).listSensors(), throwsException);
  });

  test('rejects an invalid JSON response', () async {
    final client = MockClient((_) async => http.Response('not-json', 200));

    expect(SensorApi(client).listSensors(), throwsFormatException);
  });

  test('loads readings with a UTC time window', () async {
    final client = MockClient((request) async {
      expect(request.url.path, '/api/sensors/nox-analyzer-1/readings');
      expect(request.url.queryParameters['from'], '2026-07-20T14:00:00.000Z');
      expect(request.url.queryParameters['to'], '2026-07-20T15:00:00.000Z');
      return http.Response('''
        [{
          "recorded_at": "2026-07-20T08:03:00-06:00",
          "value": 41.2,
          "status": "valid"
        }]
      ''', 200);
    });

    final readings = await SensorApi(client).listReadings(
      'nox-analyzer-1',
      DateTime.parse('2026-07-20T08:00:00-06:00'),
      DateTime.parse('2026-07-20T09:00:00-06:00'),
    );

    expect(readings, hasLength(1));
    expect(
      readings.single.recordedAt.toUtc(),
      DateTime.parse('2026-07-20T14:03:00Z'),
    );
    expect(readings.single.value, 41.2);
  });

  test('returns an empty reading list', () async {
    final client = MockClient((_) async => http.Response('[]', 200));

    final readings = await SensorApi(client).listReadings(
      'nox-analyzer-1',
      DateTime.parse('2026-07-20T14:00:00Z'),
      DateTime.parse('2026-07-20T15:00:00Z'),
    );

    expect(readings, isEmpty);
  });

  test('rejects a non-success reading response', () async {
    final client = MockClient((_) async => http.Response('unavailable', 503));
    final api = SensorApi(client);

    expect(
      api.listReadings(
        'nox-analyzer-1',
        DateTime.parse('2026-07-20T14:00:00Z'),
        DateTime.parse('2026-07-20T15:00:00Z'),
      ),
      throwsException,
    );
  });

  test('rejects a non-array reading response', () async {
    final client = MockClient((_) async => http.Response('{}', 200));
    final api = SensorApi(client);

    expect(
      api.listReadings(
        'nox-analyzer-1',
        DateTime.parse('2026-07-20T14:00:00Z'),
        DateTime.parse('2026-07-20T15:00:00Z'),
      ),
      throwsFormatException,
    );
  });

  test('loads hourly summaries with a UTC time window', () async {
    final client = MockClient((request) async {
      expect(request.url.path, '/api/sensors/nox-analyzer-1/summary');
      expect(request.url.queryParameters['from'], '2026-07-20T14:00:00.000Z');
      expect(request.url.queryParameters['to'], '2026-07-20T16:00:00.000Z');
      return http.Response('''
        [{
          "bucket_start": "2026-07-20T08:00:00-06:00",
          "average": 41.2,
          "minimum": 40.9,
          "maximum": 41.5,
          "valid_count": 2,
          "out_of_range_count": 1
        }]
      ''', 200);
    });

    final summaries = await SensorApi(client).listSummaries(
      'nox-analyzer-1',
      DateTime.parse('2026-07-20T08:00:00-06:00'),
      DateTime.parse('2026-07-20T10:00:00-06:00'),
    );

    expect(summaries, hasLength(1));
    expect(
      summaries.single.bucketStart.toUtc(),
      DateTime.parse('2026-07-20T14:00:00Z'),
    );
    expect(summaries.single.validCount, 2);
  });

  test('returns an empty summary list', () async {
    final client = MockClient((_) async => http.Response('[]', 200));

    final summaries = await SensorApi(client).listSummaries(
      'nox-analyzer-1',
      DateTime.parse('2026-07-20T14:00:00Z'),
      DateTime.parse('2026-07-20T15:00:00Z'),
    );

    expect(summaries, isEmpty);
  });

  test('rejects a non-success summary response', () async {
    final client = MockClient((_) async => http.Response('unavailable', 503));
    final api = SensorApi(client);

    expect(
      api.listSummaries(
        'nox-analyzer-1',
        DateTime.parse('2026-07-20T14:00:00Z'),
        DateTime.parse('2026-07-20T15:00:00Z'),
      ),
      throwsException,
    );
  });

  test('rejects a non-array summary response', () async {
    final client = MockClient((_) async => http.Response('{}', 200));
    final api = SensorApi(client);

    expect(
      api.listSummaries(
        'nox-analyzer-1',
        DateTime.parse('2026-07-20T14:00:00Z'),
        DateTime.parse('2026-07-20T15:00:00Z'),
      ),
      throwsFormatException,
    );
  });
}
