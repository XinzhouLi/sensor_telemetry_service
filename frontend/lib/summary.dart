class SummaryBucket {
  const SummaryBucket({
    required this.bucketStart,
    required this.average,
    required this.minimum,
    required this.maximum,
    required this.validCount,
    required this.outOfRangeCount,
  });

  factory SummaryBucket.fromJson(Map<String, dynamic> json) {
    return SummaryBucket(
      bucketStart: DateTime.parse(json['bucket_start'] as String),
      average: (json['average'] as num?)?.toDouble(),
      minimum: (json['minimum'] as num?)?.toDouble(),
      maximum: (json['maximum'] as num?)?.toDouble(),
      validCount: (json['valid_count'] as num).toInt(),
      outOfRangeCount: (json['out_of_range_count'] as num).toInt(),
    );
  }

  final DateTime bucketStart;
  final double? average;
  final double? minimum;
  final double? maximum;
  final int validCount;
  final int outOfRangeCount;
}
