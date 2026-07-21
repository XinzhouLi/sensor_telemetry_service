class Sensor {
  const Sensor({
    required this.id,
    required this.name,
    required this.unit,
    required this.health,
    this.latestReading,
  });

  factory Sensor.fromJson(Map<String, dynamic> json) {
    final latestReading = json['latest_reading'];
    return Sensor(
      id: json['id'] as String,
      name: json['name'] as String,
      unit: json['unit'] as String,
      health: json['health'] as String,
      latestReading: latestReading == null
          ? null
          : LatestReading.fromJson(latestReading as Map<String, dynamic>),
    );
  }

  final String id;
  final String name;
  final String unit;
  final String health;
  final LatestReading? latestReading;
}

class LatestReading {
  const LatestReading({
    required this.recordedAt,
    required this.value,
    required this.status,
  });

  factory LatestReading.fromJson(Map<String, dynamic> json) {
    return LatestReading(
      recordedAt: DateTime.parse(json['recorded_at'] as String),
      value: (json['value'] as num).toDouble(),
      status: json['status'] as String,
    );
  }

  final DateTime recordedAt;
  final double value;
  final String status;
}
