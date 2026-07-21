import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;

import 'dashboard_page.dart';
import 'sensor.dart';
import 'sensor_api.dart';

void main() {
  final api = SensorApi(http.Client());
  runApp(SensorDashboardApp(loadSensors: api.listSensors));
}

class SensorDashboardApp extends StatelessWidget {
  const SensorDashboardApp({super.key, required this.loadSensors});

  final Future<List<Sensor>> Function() loadSensors;

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      debugShowCheckedModeBanner: false,
      title: 'Sensor Telemetry Dashboard',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.blueGrey),
        useMaterial3: true,
      ),
      home: DashboardPage(loadSensors: loadSensors),
    );
  }
}
