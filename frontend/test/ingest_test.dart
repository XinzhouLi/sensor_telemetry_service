import 'package:flutter_test/flutter_test.dart';
import 'package:sensor_dashboard/ingest.dart';

void main() {
  test('parses stored and conflict ingestion fields', () {
    final stored = IngestResponse.fromJson({
      'stored': 1,
      'duplicates': 0,
      'conflicts': 0,
      'rejected': 0,
      'results': [
        {'index': 0, 'outcome': 'stored', 'status': 'valid'},
      ],
    });
    final conflict = IngestResult.fromJson({
      'index': 0,
      'outcome': 'conflict',
      'status': 'valid',
      'existing_value': 41.2,
    });

    expect(stored.stored, 1);
    expect(stored.results.single.status, 'valid');
    expect(conflict.outcome, 'conflict');
    expect(conflict.existingValue, 41.2);
  });

  test('parses a rejected ingestion result', () {
    final result = IngestResult.fromJson({
      'index': 0,
      'outcome': 'rejected',
      'error': 'value is required',
    });

    expect(result.status, isNull);
    expect(result.error, 'value is required');
  });
}
