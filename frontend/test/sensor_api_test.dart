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
}
