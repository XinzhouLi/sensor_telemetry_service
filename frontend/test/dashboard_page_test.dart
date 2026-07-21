import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:sensor_dashboard/dashboard_page.dart';
import 'package:sensor_dashboard/sensor.dart';

void main() {
  testWidgets('shows a loading indicator while sensors are loading', (
    tester,
  ) async {
    addTearDown(() => tester.pumpWidget(const SizedBox()));
    final result = Completer<List<Sensor>>();
    await tester.pumpWidget(_app(() => result.future));

    expect(find.byType(CircularProgressIndicator), findsOneWidget);

    result.complete([]);
    await tester.pumpAndSettle();
  });

  testWidgets('shows sensor reading and health details', (tester) async {
    addTearDown(() => tester.pumpWidget(const SizedBox()));
    await tester.pumpWidget(_app(() async => [_reportedSensor()]));
    await tester.pumpAndSettle();

    expect(find.text('NOx Analyzer 1'), findsOneWidget);
    expect(find.text('nox-analyzer-1'), findsOneWidget);
    expect(find.text('41.2 ppm'), findsOneWidget);
    expect(find.text('valid'), findsOneWidget);
    expect(find.text('ok'), findsOneWidget);
    expect(find.text('2026-07-20T14:03:00.000Z'), findsOneWidget);
  });

  testWidgets('shows a sensor without a latest reading', (tester) async {
    addTearDown(() => tester.pumpWidget(const SizedBox()));
    const sensor = Sensor(
      id: 'stack-temp-1',
      name: 'Stack Temperature 1',
      unit: '°C',
      health: 'never_reported',
    );
    await tester.pumpWidget(_app(() async => [sensor]));
    await tester.pumpAndSettle();

    expect(find.text('No readings'), findsOneWidget);
    expect(find.text('never_reported'), findsOneWidget);
  });

  testWidgets('shows an empty state when no sensors exist', (tester) async {
    addTearDown(() => tester.pumpWidget(const SizedBox()));
    await tester.pumpWidget(_app(() async => []));
    await tester.pumpAndSettle();

    expect(find.text('No sensors found.'), findsOneWidget);
  });

  testWidgets('shows an error and retries the request', (tester) async {
    addTearDown(() => tester.pumpWidget(const SizedBox()));
    var calls = 0;
    Future<List<Sensor>> loadSensors() async {
      calls++;
      if (calls == 1) {
        throw Exception('database unavailable');
      }
      return [_reportedSensor()];
    }

    await tester.pumpWidget(_app(loadSensors));
    await tester.pumpAndSettle();

    expect(find.text('Could not load sensors.'), findsOneWidget);
    expect(find.text('Retry'), findsOneWidget);

    await tester.tap(find.text('Retry'));
    await tester.pumpAndSettle();

    expect(calls, 2);
    expect(find.text('NOx Analyzer 1'), findsOneWidget);
    expect(find.text('Could not load sensors.'), findsNothing);
  });

  testWidgets('refreshes sensors every 15 seconds', (tester) async {
    addTearDown(() => tester.pumpWidget(const SizedBox()));
    var calls = 0;

    await tester.pumpWidget(
      _app(() async {
        calls++;
        return [_reportedSensor()];
      }),
    );
    await tester.pumpAndSettle();

    expect(calls, 1);

    await tester.pump(const Duration(seconds: 15));
    await tester.pump();

    expect(calls, 2);
  });
}

Widget _app(Future<List<Sensor>> Function() loadSensors) {
  return MaterialApp(home: DashboardPage(loadSensors: loadSensors));
}

Sensor _reportedSensor() {
  return Sensor(
    id: 'nox-analyzer-1',
    name: 'NOx Analyzer 1',
    unit: 'ppm',
    health: 'ok',
    latestReading: LatestReading(
      recordedAt: DateTime.parse('2026-07-20T08:03:00-06:00'),
      value: 41.2,
      status: 'valid',
    ),
  );
}
