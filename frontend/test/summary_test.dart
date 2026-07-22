import 'package:flutter_test/flutter_test.dart';
import 'package:sensor_dashboard/summary.dart';

void main() {
  test('parses a summary bucket with statistics', () {
    final bucket = SummaryBucket.fromJson({
      'bucket_start': '2026-07-20T08:00:00-06:00',
      'average': 41.2,
      'minimum': 40.9,
      'maximum': 41.5,
      'valid_count': 2,
      'out_of_range_count': 1,
    });

    expect(bucket.bucketStart.toUtc(), DateTime.parse('2026-07-20T14:00:00Z'));
    expect(bucket.average, 41.2);
    expect(bucket.minimum, 40.9);
    expect(bucket.maximum, 41.5);
    expect(bucket.validCount, 2);
    expect(bucket.outOfRangeCount, 1);
  });

  test('parses null statistics when a bucket has no valid readings', () {
    final bucket = SummaryBucket.fromJson({
      'bucket_start': '2026-07-20T15:00:00Z',
      'average': null,
      'minimum': null,
      'maximum': null,
      'valid_count': 0,
      'out_of_range_count': 2,
    });

    expect(bucket.average, isNull);
    expect(bucket.minimum, isNull);
    expect(bucket.maximum, isNull);
    expect(bucket.validCount, 0);
    expect(bucket.outOfRangeCount, 2);
  });
}
