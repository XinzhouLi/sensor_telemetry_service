class IngestResponse {
  const IngestResponse({
    required this.stored,
    required this.duplicates,
    required this.conflicts,
    required this.rejected,
    required this.results,
  });

  factory IngestResponse.fromJson(Map<String, dynamic> json) {
    return IngestResponse(
      stored: (json['stored'] as num).toInt(),
      duplicates: (json['duplicates'] as num).toInt(),
      conflicts: (json['conflicts'] as num).toInt(),
      rejected: (json['rejected'] as num).toInt(),
      results: (json['results'] as List)
          .map((item) => IngestResult.fromJson(item as Map<String, dynamic>))
          .toList(),
    );
  }

  final int stored;
  final int duplicates;
  final int conflicts;
  final int rejected;
  final List<IngestResult> results;
}

class IngestResult {
  const IngestResult({
    required this.index,
    required this.outcome,
    this.status,
    this.existingValue,
    this.error,
  });

  factory IngestResult.fromJson(Map<String, dynamic> json) {
    return IngestResult(
      index: (json['index'] as num).toInt(),
      outcome: json['outcome'] as String,
      status: json['status'] as String?,
      existingValue: (json['existing_value'] as num?)?.toDouble(),
      error: json['error'] as String?,
    );
  }

  final int index;
  final String outcome;
  final String? status;
  final double? existingValue;
  final String? error;
}
