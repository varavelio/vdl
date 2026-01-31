// ignore_for_file: unused_element, unused_field

/// Response represents the result of a VDL RPC call.
///
/// A response is either successful ([ok] is `true`) with an [output] value,
/// or failed ([ok] is `false`) with an [error] describing what went wrong.
class Response<T> {
  /// Indicates whether the RPC call was successful.
  final bool ok;

  /// The successful output (only present if [ok] is `true`).
  final T? output;

  /// Structured error (only present if [ok] is `false`).
  final VdlError? error;

  const Response._({required this.ok, this.output, this.error});

  /// Creates a successful response with the given [output].
  factory Response.ok(T output) => Response._(ok: true, output: output);

  /// Creates a failed response with the given [error].
  factory Response.error(VdlError error) => Response._(ok: false, error: error);

  /// Creates a [Response] from a JSON map.
  ///
  /// The [hydrateOutput] function is used to convert the raw output to type [T].
  /// If not provided, the output is cast directly to [T].
  factory Response.fromJson(
    Map<String, dynamic> json, {
    T Function(Map<String, dynamic>)? hydrateOutput,
  }) {
    if (json['ok'] == true) {
      final rawOutput = json['output'];
      if (hydrateOutput != null && rawOutput is Map<String, dynamic>) {
        return Response.ok(hydrateOutput(rawOutput));
      }
      return Response.ok(rawOutput as T);
    }
    final err = json['error'];
    return Response.error(
      err is Map<String, dynamic>
          ? VdlError.fromJson(err)
          : VdlError(message: err?.toString() ?? 'Unknown error'),
    );
  }

  /// Converts this [Response] to a JSON map.
  Map<String, dynamic> toJson() {
    if (ok) {
      return {'ok': true, 'output': output};
    }
    return {'ok': false, 'error': error?.toJson()};
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) return true;
    return other is Response<T> &&
        ok == other.ok &&
        output == other.output &&
        error == other.error;
  }

  @override
  int get hashCode => Object.hash(ok, output, error);

  @override
  String toString() {
    if (ok) {
      return 'Response.ok($output)';
    }
    return 'Response.error($error)';
  }
}

/// Structured error type used throughout the VDL ecosystem.
///
/// A [VdlError] provides detailed information about what went wrong,
/// including an optional [category] for grouping errors, a [code] for
/// programmatic handling, and [details] for additional context.
class VdlError implements Exception {
  /// Human-readable description of the error.
  final String message;

  /// Categorizes the error by its nature or source (e.g., "ValidationError").
  final String? category;

  /// Machine-readable identifier for the specific error condition.
  final String? code;

  /// Additional information about the error.
  final Map<String, dynamic>? details;

  /// Creates a new [VdlError] with the given [message] and optional fields.
  const VdlError({
    required this.message,
    this.category,
    this.code,
    this.details,
  });

  /// Creates a [VdlError] from a JSON map.
  factory VdlError.fromJson(Map<String, dynamic> json) => VdlError(
    message: json['message']?.toString() ?? 'Unknown error',
    category: json['category'] as String?,
    code: json['code'] as String?,
    details: json['details'] is Map
        ? (json['details'] as Map).cast<String, dynamic>()
        : null,
  );

  /// Converts this [VdlError] to a JSON map.
  Map<String, dynamic> toJson() {
    final data = <String, dynamic>{'message': message};
    if (category != null) data['category'] = category;
    if (code != null) data['code'] = code;
    if (details != null) data['details'] = details;
    return data;
  }

  /// Creates a copy of this [VdlError] with the given fields replaced.
  VdlError copyWith({
    String? message,
    String? category,
    String? code,
    Map<String, dynamic>? details,
  }) {
    return VdlError(
      message: message ?? this.message,
      category: category ?? this.category,
      code: code ?? this.code,
      details: details ?? this.details,
    );
  }

  @override
  bool operator ==(Object other) {
    if (identical(this, other)) return true;
    return other is VdlError &&
        message == other.message &&
        category == other.category &&
        code == other.code;
  }

  @override
  int get hashCode => Object.hash(message, category, code);

  @override
  String toString() {
    final parts = <String>['message: $message'];
    if (category != null) parts.add('category: $category');
    if (code != null) parts.add('code: $code');
    return 'VdlError(${parts.join(', ')})';
  }
}
