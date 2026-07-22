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

  testWidgets('requires a sensor selection before showing query controls', (
    tester,
  ) async {
    addTearDown(() => tester.pumpWidget(const SizedBox()));
    await tester.pumpWidget(_app(() async => [_reportedSensor()]));
    await tester.pumpAndSettle();

    expect(find.text('Select a sensor to view readings.'), findsOneWidget);
    final button = tester.widget<FilledButton>(
      find.byKey(const Key('query-readings')),
    );
    expect(button.onPressed, isNull);
  });

  testWidgets('selects a sensor and starts with a 24 hour UTC window', (
    tester,
  ) async {
    addTearDown(() => tester.pumpWidget(const SizedBox()));
    await tester.pumpWidget(
      _app(
        () async => [_reportedSensor()],
        now: () => DateTime.parse('2026-07-21T15:00:00Z'),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.byKey(const Key('sensor-card-nox-analyzer-1')));
    await tester.pump();

    expect(
      _field(tester, const Key('from-field')).controller!.text,
      '2026-07-20T15:00:00.000Z',
    );
    expect(
      _field(tester, const Key('to-field')).controller!.text,
      '2026-07-21T15:00:00.000Z',
    );
  });

  testWidgets('rejects invalid reading windows without calling the API', (
    tester,
  ) async {
    addTearDown(() => tester.pumpWidget(const SizedBox()));
    var calls = 0;
    await tester.pumpWidget(
      _app(
        () async => [_reportedSensor()],
        loadReadings: (_, _, _) async {
          calls++;
          return [];
        },
      ),
    );
    await tester.pumpAndSettle();
    await _selectSensor(tester, 'nox-analyzer-1');

    final from = find.byKey(const Key('from-field'));
    final to = find.byKey(const Key('to-field'));
    final query = find.byKey(const Key('query-readings'));

    await tester.enterText(from, '');
    await tester.tap(query);
    await tester.pump();
    expect(find.text('From and to are required.'), findsOneWidget);

    await tester.enterText(from, 'not-a-time');
    await tester.tap(query);
    await tester.pump();
    expect(
      find.text('From must be a valid RFC3339 timestamp.'),
      findsOneWidget,
    );

    await tester.enterText(from, '2026-07-21T15:00:00Z');
    await tester.enterText(to, 'not-a-time');
    await tester.tap(query);
    await tester.pump();
    expect(find.text('To must be a valid RFC3339 timestamp.'), findsOneWidget);

    await tester.enterText(to, '2026-07-21T15:00:00Z');
    await tester.tap(query);
    await tester.pump();
    expect(find.text('From must be before to.'), findsOneWidget);

    await tester.enterText(from, '2026-07-21T16:00:00Z');
    await tester.tap(query);
    await tester.pump();
    expect(find.text('From must be before to.'), findsOneWidget);
    expect(calls, 0);
  });

  testWidgets('queries offset times and renders readings in UTC', (
    tester,
  ) async {
    addTearDown(() => tester.pumpWidget(const SizedBox()));
    String? sensorID;
    DateTime? receivedFrom;
    DateTime? receivedTo;
    await tester.pumpWidget(
      _app(
        () async => [_reportedSensor()],
        loadReadings: (id, from, to) async {
          sensorID = id;
          receivedFrom = from;
          receivedTo = to;
          return [
            Reading(
              recordedAt: DateTime.parse('2026-07-20T08:03:00-06:00'),
              value: 41.2,
              status: 'valid',
            ),
            Reading(
              recordedAt: DateTime.parse('2026-07-20T14:04:00Z'),
              value: 512,
              status: 'out_of_range',
            ),
          ];
        },
      ),
    );
    await tester.pumpAndSettle();
    await _selectSensor(tester, 'nox-analyzer-1');
    await tester.enterText(
      find.byKey(const Key('from-field')),
      '2026-07-20T08:00:00-06:00',
    );
    await tester.enterText(
      find.byKey(const Key('to-field')),
      '2026-07-20T09:00:00-06:00',
    );

    await tester.tap(find.byKey(const Key('query-readings')));
    await tester.pumpAndSettle();

    expect(sensorID, 'nox-analyzer-1');
    expect(receivedFrom!.toUtc(), DateTime.parse('2026-07-20T14:00:00Z'));
    expect(receivedTo!.toUtc(), DateTime.parse('2026-07-20T15:00:00Z'));
    expect(find.text('2026-07-20T14:03:00.000Z'), findsWidgets);
    expect(find.text('41.2 ppm'), findsWidgets);
    expect(find.text('valid'), findsWidgets);
    expect(find.text('512.0 ppm'), findsOneWidget);
    expect(find.text('out_of_range'), findsOneWidget);
  });

  testWidgets('disables Query while readings are loading', (tester) async {
    addTearDown(() => tester.pumpWidget(const SizedBox()));
    final result = Completer<List<Reading>>();
    await tester.pumpWidget(
      _app(
        () async => [_reportedSensor()],
        loadReadings: (_, _, _) => result.future,
      ),
    );
    await tester.pumpAndSettle();
    await _selectSensor(tester, 'nox-analyzer-1');

    await tester.tap(find.byKey(const Key('query-readings')));
    await tester.pump();

    final button = tester.widget<FilledButton>(
      find.byKey(const Key('query-readings')),
    );
    expect(button.onPressed, isNull);
    expect(find.byType(CircularProgressIndicator), findsOneWidget);

    result.complete([]);
    await tester.pumpAndSettle();
  });

  testWidgets('shows empty and error reading results', (tester) async {
    addTearDown(() => tester.pumpWidget(const SizedBox()));
    var fail = false;
    await tester.pumpWidget(
      _app(
        () async => [_reportedSensor()],
        loadReadings: (_, _, _) async {
          if (fail) {
            throw Exception('database unavailable');
          }
          return [];
        },
      ),
    );
    await tester.pumpAndSettle();
    await _selectSensor(tester, 'nox-analyzer-1');

    await tester.tap(find.byKey(const Key('query-readings')));
    await tester.pumpAndSettle();
    expect(find.text('No readings in this time window.'), findsOneWidget);

    fail = true;
    await tester.tap(find.byKey(const Key('query-readings')));
    await tester.pumpAndSettle();
    expect(find.text('Could not load readings.'), findsOneWidget);
  });

  testWidgets('switching sensors clears results but keeps the time window', (
    tester,
  ) async {
    addTearDown(() => tester.pumpWidget(const SizedBox()));
    final secondSensor = Sensor(
      id: 'o2-analyzer-1',
      name: 'Oxygen Analyzer 1',
      unit: '%',
      health: 'stale',
      latestReading: Reading(
        recordedAt: DateTime.parse('2026-07-20T14:00:00Z'),
        value: 20.8,
        status: 'valid',
      ),
    );
    await tester.pumpWidget(
      _app(
        () async => [_reportedSensor(), secondSensor],
        loadReadings: (_, _, _) async => [
          Reading(
            recordedAt: DateTime.parse('2026-07-20T14:03:00Z'),
            value: 41.2,
            status: 'valid',
          ),
        ],
      ),
    );
    await tester.pumpAndSettle();
    await _selectSensor(tester, 'nox-analyzer-1');
    await tester.enterText(
      find.byKey(const Key('from-field')),
      '2026-07-20T14:00:00Z',
    );
    await tester.enterText(
      find.byKey(const Key('to-field')),
      '2026-07-20T15:00:00Z',
    );
    await tester.tap(find.byKey(const Key('query-readings')));
    await tester.pumpAndSettle();
    expect(find.text('41.2 ppm'), findsWidgets);

    await _selectSensor(tester, 'o2-analyzer-1');

    expect(find.text('Recorded at (UTC)'), findsNothing);
    expect(
      _field(tester, const Key('from-field')).controller!.text,
      '2026-07-20T14:00:00Z',
    );
    expect(
      _field(tester, const Key('to-field')).controller!.text,
      '2026-07-20T15:00:00Z',
    );
  });

  testWidgets('overview refresh does not query readings', (tester) async {
    addTearDown(() => tester.pumpWidget(const SizedBox()));
    var readingCalls = 0;
    await tester.pumpWidget(
      _app(
        () async => [_reportedSensor()],
        loadReadings: (_, _, _) async {
          readingCalls++;
          return [];
        },
      ),
    );
    await tester.pumpAndSettle();
    await _selectSensor(tester, 'nox-analyzer-1');

    await tester.pump(const Duration(seconds: 15));
    await tester.pump();

    expect(readingCalls, 0);
  });
}

Widget _app(
  Future<List<Sensor>> Function() loadSensors, {
  Future<List<Reading>> Function(String, DateTime, DateTime)? loadReadings,
  DateTime Function() now = DateTime.now,
}) {
  return MaterialApp(
    home: DashboardPage(
      loadSensors: loadSensors,
      loadReadings: loadReadings ?? (_, _, _) async => [],
      now: now,
    ),
  );
}

Future<void> _selectSensor(WidgetTester tester, String sensorID) async {
  await tester.tap(find.byKey(Key('sensor-card-$sensorID')));
  await tester.pump();
  await tester.ensureVisible(find.byKey(const Key('query-readings')));
  await tester.pump();
}

TextField _field(WidgetTester tester, Key key) {
  return tester.widget<TextField>(find.byKey(key));
}

Sensor _reportedSensor() {
  return Sensor(
    id: 'nox-analyzer-1',
    name: 'NOx Analyzer 1',
    unit: 'ppm',
    health: 'ok',
    latestReading: Reading(
      recordedAt: DateTime.parse('2026-07-20T08:03:00-06:00'),
      value: 41.2,
      status: 'valid',
    ),
  );
}
