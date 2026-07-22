import 'dart:async';

import 'package:flutter/material.dart';

import 'sensor.dart';

class DashboardPage extends StatefulWidget {
  const DashboardPage({
    super.key,
    required this.loadSensors,
    required this.loadReadings,
    this.now = DateTime.now,
  });

  final Future<List<Sensor>> Function() loadSensors;
  final Future<List<Reading>> Function(String, DateTime, DateTime) loadReadings;
  final DateTime Function() now;

  @override
  State<DashboardPage> createState() => _DashboardPageState();
}

class _DashboardPageState extends State<DashboardPage> {
  static const _refreshInterval = Duration(seconds: 15);

  List<Sensor>? _sensors;
  bool _failed = false;
  Timer? _refreshTimer;
  late final TextEditingController _fromController;
  late final TextEditingController _toController;
  String? _selectedSensorID;
  List<Reading>? _readings;
  bool _readingsLoading = false;
  bool _hasQueriedReadings = false;
  String? _readingsError;
  int _readingRequestID = 0;

  @override
  void initState() {
    super.initState();
    final to = widget.now().toUtc();
    _fromController = TextEditingController(
      text: to.subtract(const Duration(hours: 24)).toIso8601String(),
    );
    _toController = TextEditingController(text: to.toIso8601String());
    _load();
    _refreshTimer = Timer.periodic(_refreshInterval, (_) => _load());
  }

  @override
  void dispose() {
    _refreshTimer?.cancel();
    _fromController.dispose();
    _toController.dispose();
    super.dispose();
  }

  Future<void> _load() async {
    setState(() {
      _failed = false;
    });

    try {
      final sensors = await widget.loadSensors();
      if (!mounted) {
        return;
      }
      setState(() {
        _sensors = sensors;
        if (_selectedSensorID != null &&
            !sensors.any((sensor) => sensor.id == _selectedSensorID)) {
          _selectedSensorID = null;
          _clearReadingResult();
        }
      });
    } catch (_) {
      if (!mounted) {
        return;
      }
      setState(() {
        _failed = true;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Sensor Telemetry Dashboard')),
      body: _buildBody(),
    );
  }

  Widget _buildBody() {
    if (_failed) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Text('Could not load sensors.'),
            const SizedBox(height: 12),
            FilledButton(onPressed: _load, child: const Text('Retry')),
          ],
        ),
      );
    }
    if (_sensors == null) {
      return const Center(child: CircularProgressIndicator());
    }
    if (_sensors!.isEmpty) {
      return const Center(child: Text('No sensors found.'));
    }

    return ListView(
      padding: const EdgeInsets.all(24),
      children: [
        GridView.builder(
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          gridDelegate: const SliverGridDelegateWithMaxCrossAxisExtent(
            maxCrossAxisExtent: 420,
            mainAxisExtent: 280,
            crossAxisSpacing: 16,
            mainAxisSpacing: 16,
          ),
          itemCount: _sensors!.length,
          itemBuilder: (context, index) {
            final sensor = _sensors![index];
            return _SensorCard(
              sensor: sensor,
              selected: sensor.id == _selectedSensorID,
              onTap: () => _selectSensor(sensor.id),
            );
          },
        ),
        const SizedBox(height: 24),
        _buildReadingSection(),
      ],
    );
  }

  void _selectSensor(String sensorID) {
    if (sensorID == _selectedSensorID) {
      return;
    }
    setState(() {
      _selectedSensorID = sensorID;
      _readingRequestID++;
      _clearReadingResult();
    });
  }

  void _clearReadingResult() {
    _readings = null;
    _readingsLoading = false;
    _hasQueriedReadings = false;
    _readingsError = null;
  }

  Sensor? get _selectedSensor {
    for (final sensor in _sensors ?? const <Sensor>[]) {
      if (sensor.id == _selectedSensorID) {
        return sensor;
      }
    }
    return null;
  }

  Future<void> _queryReadings() async {
    final sensor = _selectedSensor;
    if (sensor == null || _readingsLoading) {
      return;
    }

    final fromText = _fromController.text.trim();
    final toText = _toController.text.trim();
    if (fromText.isEmpty || toText.isEmpty) {
      setState(() => _readingsError = 'From and to are required.');
      return;
    }

    final from = _parseRFC3339(fromText);
    if (from == null) {
      setState(
        () => _readingsError = 'From must be a valid RFC3339 timestamp.',
      );
      return;
    }
    final to = _parseRFC3339(toText);
    if (to == null) {
      setState(() => _readingsError = 'To must be a valid RFC3339 timestamp.');
      return;
    }
    if (!from.isBefore(to)) {
      setState(() => _readingsError = 'From must be before to.');
      return;
    }

    final requestID = ++_readingRequestID;
    setState(() {
      _readingsLoading = true;
      _hasQueriedReadings = true;
      _readings = null;
      _readingsError = null;
    });

    try {
      final readings = await widget.loadReadings(sensor.id, from, to);
      if (!mounted || requestID != _readingRequestID) {
        return;
      }
      setState(() {
        _readings = readings;
        _readingsLoading = false;
      });
    } catch (_) {
      if (!mounted || requestID != _readingRequestID) {
        return;
      }
      setState(() {
        _readingsLoading = false;
        _readingsError = 'Could not load readings.';
      });
    }
  }

  Widget _buildReadingSection() {
    final sensor = _selectedSensor;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Reading window',
              style: Theme.of(context).textTheme.titleLarge,
            ),
            const SizedBox(height: 16),
            if (sensor == null)
              const Text('Select a sensor to view readings.')
            else
              Text(sensor.name, style: Theme.of(context).textTheme.titleMedium),
            const SizedBox(height: 16),
            Wrap(
              spacing: 12,
              runSpacing: 12,
              crossAxisAlignment: WrapCrossAlignment.center,
              children: [
                SizedBox(
                  width: 320,
                  child: TextField(
                    key: const Key('from-field'),
                    controller: _fromController,
                    decoration: const InputDecoration(
                      border: OutlineInputBorder(),
                      labelText: 'From (RFC3339 UTC)',
                    ),
                  ),
                ),
                SizedBox(
                  width: 320,
                  child: TextField(
                    key: const Key('to-field'),
                    controller: _toController,
                    decoration: const InputDecoration(
                      border: OutlineInputBorder(),
                      labelText: 'To (RFC3339 UTC)',
                    ),
                  ),
                ),
                FilledButton(
                  key: const Key('query-readings'),
                  onPressed: sensor == null || _readingsLoading
                      ? null
                      : _queryReadings,
                  child: const Text('Query'),
                ),
              ],
            ),
            if (sensor != null) ...[
              const SizedBox(height: 16),
              if (_readingsError != null)
                Text(
                  _readingsError!,
                  key: const Key('readings-error'),
                  style: TextStyle(color: Theme.of(context).colorScheme.error),
                )
              else if (_readingsLoading)
                const Center(child: CircularProgressIndicator())
              else if (_hasQueriedReadings && _readings!.isEmpty)
                const Text('No readings in this time window.')
              else if (_readings != null)
                SingleChildScrollView(
                  scrollDirection: Axis.horizontal,
                  child: DataTable(
                    columns: const [
                      DataColumn(label: Text('Recorded at (UTC)')),
                      DataColumn(label: Text('Value')),
                      DataColumn(label: Text('Status')),
                    ],
                    rows: _readings!
                        .map(
                          (reading) => DataRow(
                            cells: [
                              DataCell(
                                Text(
                                  reading.recordedAt.toUtc().toIso8601String(),
                                ),
                              ),
                              DataCell(Text('${reading.value} ${sensor.unit}')),
                              DataCell(Text(reading.status)),
                            ],
                          ),
                        )
                        .toList(),
                  ),
                ),
            ],
          ],
        ),
      ),
    );
  }
}

final _rfc3339Pattern = RegExp(
  r'^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})$',
);

DateTime? _parseRFC3339(String text) {
  if (!_rfc3339Pattern.hasMatch(text)) {
    return null;
  }
  return DateTime.tryParse(text);
}

class _SensorCard extends StatelessWidget {
  const _SensorCard({
    required this.sensor,
    required this.selected,
    required this.onTap,
  });

  final Sensor sensor;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final reading = sensor.latestReading;
    return Card(
      key: Key('sensor-card-${sensor.id}'),
      clipBehavior: Clip.antiAlias,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: selected
            ? BorderSide(color: Theme.of(context).colorScheme.primary, width: 2)
            : BorderSide.none,
      ),
      child: InkWell(
        onTap: onTap,
        child: Padding(
          padding: const EdgeInsets.all(20),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Expanded(
                    child: Text(
                      sensor.name,
                      style: Theme.of(context).textTheme.titleLarge,
                    ),
                  ),
                  const SizedBox(width: 12),
                  _Badge(
                    label: sensor.health,
                    color: _healthColor(sensor.health),
                  ),
                ],
              ),
              Text(sensor.id, style: Theme.of(context).textTheme.bodySmall),
              const Divider(height: 32),
              if (reading == null)
                const Expanded(child: Center(child: Text('No readings')))
              else ...[
                Text(
                  '${reading.value} ${sensor.unit}',
                  style: Theme.of(context).textTheme.headlineMedium,
                ),
                const SizedBox(height: 12),
                _Badge(
                  label: reading.status,
                  color: _statusColor(reading.status),
                ),
                const Spacer(),
                Text(
                  reading.recordedAt.toUtc().toIso8601String(),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: Theme.of(context).textTheme.bodySmall,
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}

class _Badge extends StatelessWidget {
  const _Badge({required this.label, required this.color});

  final String label;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.14),
        borderRadius: BorderRadius.circular(16),
      ),
      child: Text(
        label,
        style: TextStyle(color: color, fontWeight: FontWeight.w600),
      ),
    );
  }
}

Color _healthColor(String health) {
  return switch (health) {
    'ok' => Colors.green.shade700,
    'stale' => Colors.orange.shade800,
    'never_reported' => Colors.blueGrey.shade600,
    _ => Colors.blueGrey.shade600,
  };
}

Color _statusColor(String status) {
  return switch (status) {
    'valid' => Colors.green.shade700,
    'out_of_range' => Colors.red.shade700,
    _ => Colors.blueGrey.shade600,
  };
}
