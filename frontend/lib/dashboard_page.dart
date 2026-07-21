import 'package:flutter/material.dart';

import 'sensor.dart';

class DashboardPage extends StatefulWidget {
  const DashboardPage({super.key, required this.loadSensors});

  final Future<List<Sensor>> Function() loadSensors;

  @override
  State<DashboardPage> createState() => _DashboardPageState();
}

class _DashboardPageState extends State<DashboardPage> {
  List<Sensor>? _sensors;
  bool _failed = false;

  @override
  void initState() {
    super.initState();
    _load();
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

    return GridView.builder(
      padding: const EdgeInsets.all(24),
      gridDelegate: const SliverGridDelegateWithMaxCrossAxisExtent(
        maxCrossAxisExtent: 420,
        mainAxisExtent: 260,
        crossAxisSpacing: 16,
        mainAxisSpacing: 16,
      ),
      itemCount: _sensors!.length,
      itemBuilder: (context, index) => _SensorCard(sensor: _sensors![index]),
    );
  }
}

class _SensorCard extends StatelessWidget {
  const _SensorCard({required this.sensor});

  final Sensor sensor;

  @override
  Widget build(BuildContext context) {
    final reading = sensor.latestReading;
    return Card(
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
