import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:sensor_dashboard/ingest.dart';
import 'package:sensor_dashboard/main.dart';

void main() {
  testWidgets('renders the dashboard title', (tester) async {
    addTearDown(() => tester.pumpWidget(const SizedBox()));
    await tester.pumpWidget(
      SensorDashboardApp(
        loadSensors: () async => [],
        loadReadings: (_, _, _) async => [],
        loadSummaries: (_, _, _) async => [],
        ingestReading: (_, _, _) async => const IngestResponse(
          stored: 1,
          duplicates: 0,
          conflicts: 0,
          rejected: 0,
          results: [IngestResult(index: 0, outcome: 'stored')],
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Sensor Telemetry Dashboard'), findsOneWidget);
  });
}
