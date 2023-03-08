# this script is used to compile java and generate environment
# take input args with source file path and working directory

# check input args
if [ $# -ne 2 ]; then
    echo "Usage: $0 source_file_path working_directory"
    exit 1
fi

# check source file
if [ ! -f $1 ]; then
    echo "Source file not found"
    exit 1
fi

# check working directory
if [ ! -d $2 ]; then
    echo "Working directory not found"
    exit 1
fi

# get source file name
source_file_name=$(basename $1)
source_file_name=${source_file_name%.*}

# get source file path
source_file_path=$(dirname $1)

# get working directory
working_directory=$2

# compile java
javac -d $working_directory $source_file_path/$source_file_name.java

