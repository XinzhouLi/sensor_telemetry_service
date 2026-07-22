import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;

import 'dashboard_page.dart';
import 'sensor.dart';
import 'sensor_api.dart';
import 'summary.dart';

void main() {
  final api = SensorApi(http.Client());
  runApp(
    SensorDashboardApp(
      loadSensors: api.listSensors,
      loadReadings: api.listReadings,
      loadSummaries: api.listSummaries,
    ),
  );
}

class SensorDashboardApp extends StatelessWidget {
  const SensorDashboardApp({
    super.key,
    required this.loadSensors,
    required this.loadReadings,
    required this.loadSummaries,
  });

  final Future<List<Sensor>> Function() loadSensors;
  final Future<List<Reading>> Function(String, DateTime, DateTime) loadReadings;
  final Future<List<SummaryBucket>> Function(String, DateTime, DateTime)
  loadSummaries;

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      debugShowCheckedModeBanner: false,
      title: 'Sensor Telemetry Dashboard',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.blueGrey),
        useMaterial3: true,
      ),
      home: DashboardPage(
        loadSensors: loadSensors,
        loadReadings: loadReadings,
        loadSummaries: loadSummaries,
      ),
    );
  }
}
