# Testing `translator`

Testing of `translator` is done in two ways:
- For testing invalid IngressRoutes, add the fixtures inside `translate_test.go:TestTranslateIngressRouteErrors`.
You should probably not need to add any more of these.
- For testing valid IngressRoutes, use the `testdata` directory.

## Testdata test case structure

Each directory under the `testdata` directory is a test case, containing an `input.yaml`, an `output.yaml` and an `errors.txt` file.

The directory must contain these three files, or the test will fail.

`input.yaml` contains a YAML for an IngressRoute object, that will have the translation code run on it.
Invalid IngressRoutes here will fail the test.

`output.yaml` contains a YAML for a HTTPProxy object, as output by `ir2proxy`.

`errors.txt` should contain any warnings that should be emitted by the validation process.

If there should be no warnings, then `errors.txt` should be an empty file.

## Adding a new test case

There's a convenience script in `$REPO_ROOT/hack/newtestcase` that will create the directory and required files for you:

```sh
$REPO_ROOT $ hack/newtestcase translator newtestcase
```

will create `internal/translator/testdata/newtestcase/` and the three required files for you to fill out.

## Running the tests

Run the tests with `make check-test` from the repo root.
