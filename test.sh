go test -gcflags="all=-N -l" -c -o diff.test
# image ltarget symbols add ./mytest.testookup -r -n TestDiffStuff
# break set -n github.com/stefanpenner/c/fs-things.TestDiffStuff
# target symbols add ./mytest.test
# image list
#
# easy way: dlv test -- -test.run TestDiffStuff
