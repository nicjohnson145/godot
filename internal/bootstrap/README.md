# Running Integration Tests

Instead of some convoluted mocking, this submodule leverages Docker to actually execute the code.
Run all integration tests with

```
docker build -t godot_integration . && docker run --rm godot_integration
```

You can specify differnt build tags to to control which integration tests actually run with the
`TYPE` docker build arg.
