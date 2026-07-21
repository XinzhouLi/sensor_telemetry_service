import 'dart:convert';

import 'package:http/http.dart' as http;

import 'sensor.dart';

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
}
