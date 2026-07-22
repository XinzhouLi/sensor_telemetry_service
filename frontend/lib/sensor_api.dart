import 'dart:convert';

import 'package:http/http.dart' as http;

import 'ingest.dart';
import 'sensor.dart';
import 'summary.dart';

class SensorApi {
  SensorApi(this._client, {Uri? endpoint})
    : _endpoint = endpoint ?? Uri.parse('/api/sensors');

  final http.Client _client;
  final Uri _endpoint;

  Future<List<Sensor>> listSensors() async {
    final response = await _client.get(_endpoint);
    if (response.statusCode != 200) {
      throw Exception('failed to load sensors');
    }

    final body = jsonDecode(response.body);
    if (body is! List) {
      throw const FormatException('sensor response must be an array');
    }

    return body
        .map((item) => Sensor.fromJson(item as Map<String, dynamic>))
        .toList();
  }

  Future<List<Reading>> listReadings(
    String sensorID,
    DateTime from,
    DateTime to,
  ) async {
    final endpoint =
        Uri.parse(
          '${_endpoint.toString()}/${Uri.encodeComponent(sensorID)}/readings',
        ).replace(
          queryParameters: {
            'from': from.toUtc().toIso8601String(),
            'to': to.toUtc().toIso8601String(),
          },
        );
    final response = await _client.get(endpoint);
    if (response.statusCode != 200) {
      throw Exception('failed to load readings');
    }

    final body = jsonDecode(response.body);
    if (body is! List) {
      throw const FormatException('reading response must be an array');
    }

    return body
        .map((item) => Reading.fromJson(item as Map<String, dynamic>))
        .toList();
  }

  Future<List<SummaryBucket>> listSummaries(
    String sensorID,
    DateTime from,
    DateTime to,
  ) async {
    final endpoint =
        Uri.parse(
          '${_endpoint.toString()}/${Uri.encodeComponent(sensorID)}/summary',
        ).replace(
          queryParameters: {
            'from': from.toUtc().toIso8601String(),
            'to': to.toUtc().toIso8601String(),
          },
        );
    final response = await _client.get(endpoint);
    if (response.statusCode != 200) {
      throw Exception('failed to load summary');
    }

    final body = jsonDecode(response.body);
    if (body is! List) {
      throw const FormatException('summary response must be an array');
    }

    return body
        .map((item) => SummaryBucket.fromJson(item as Map<String, dynamic>))
        .toList();
  }

  Future<IngestResponse> ingestReading(
    String sensorID,
    DateTime recordedAt,
    double value,
  ) async {
    final endpoint = Uri.parse(
      '${_endpoint.toString()}/${Uri.encodeComponent(sensorID)}/readings',
    );
    final response = await _client.post(
      endpoint,
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode([
        {'recorded_at': recordedAt.toUtc().toIso8601String(), 'value': value},
      ]),
    );
    if (response.statusCode != 200) {
      throw Exception('failed to add reading');
    }

    final body = jsonDecode(response.body);
    if (body is! Map<String, dynamic>) {
      throw const FormatException('ingest response must be an object');
    }
    final result = IngestResponse.fromJson(body);
    if (result.results.length != 1) {
      throw const FormatException('ingest response must contain one result');
    }
    return result;
  }
}
