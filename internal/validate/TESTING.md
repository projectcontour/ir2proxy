# Testing `validate`

Testing of validate is done using the `testdata` directory.

Each directory under the `testdata` directory is a test case, containing an `input.yaml` and an `errors.txt` file.

The directory must contain these two files, or the test will fail.

`input.yaml` contains a YAML for an IngressRoute object, that will have the validation code run on it.

`errors.txt` should contain any warnings that should be emitted by the validation process.

If there should be no warnings, then `errors.txt` should be an empty file.
