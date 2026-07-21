import 'package:flutter_test/flutter_test.dart';
import 'package:sensor_dashboard/main.dart';

void main() {
  testWidgets('renders the dashboard title', (tester) async {
    await tester.pumpWidget(SensorDashboardApp(loadSensors: () async => []));
    await tester.pumpAndSettle();

    expect(find.text('Sensor Telemetry Dashboard'), findsOneWidget);
  });
}
