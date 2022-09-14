#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <sys/time.h>

#define ull unsigned long long int
#define cptr char *

ull execute_time = 0;
cptr child_result = NULL;
ull child_result_size = 0;
ull start_time = 0;
ull end_time = 0;

void append(cptr to_append, ull size) {
    if (child_result_size == 0) {
        child_result = malloc(size);
        child_result_size = size;
    } else {
        child_result = realloc(child_result, child_result_size + size);
        child_result_size += size;
    }
    for (ull i = 0; i < size; i++) {
        child_result[child_result_size - size + i] = to_append[i];
    }
}

int main(int argc, char **args) {
    if (argc < 2) {
        printf("%s", "[Usage] shell <command> ...args ##");
        return 0;
    }

    
    //create pipe for stdout
    int pipes[2] = { 0 };
    if(-1 == pipe(pipes)) {
        printf("%s", "Falied to create pipe to slave ##");
        return 0;
    }

    int pipe_stdout_read = pipes[0];
    int pipe_stdout_write = pipes[1];

    //create pipe for stdin
    if(-1 == pipe(pipes)) {
        printf("%s", "Falied to create pipe to slave ##");
        return 0;
    }

    int pipe_stdin_read = pipes[0];
    int pipe_stdin_write = pipes[1];

    //create child process
    int pid = fork();
    if (pid == 0) {
        //strict aliasing, do not allow unlimited cycles
        alarm(60);

        //child process
        //close unused pipe
        close(pipe_stdout_read);
        close(pipe_stdin_write);

        //redirect stdout
        //stdin will be read to child process, but stdout will be hijack
        dup2(pipe_stdout_write, STDOUT_FILENO);

        //execute command
        execvp(args[1], args + 1);
    } else {
        //parent process
        //close unused pipe
        close(pipe_stdout_write);
        close(pipe_stdin_read);

        //get current microsecond
        struct timeval tv;
        gettimeofday(&tv, NULL);
        start_time = tv.tv_sec * 1000000 + tv.tv_usec;

        //read from stdout
        char buf[1024] = { 0 };
        int len = 0;
        while ((len = read(pipe_stdout_read, buf, sizeof(buf))) > 0) {
            append(buf, len);
        }

        //process exit
        int status;
        waitpid(pid, &status, 0);
        //get nano time
        gettimeofday(&tv, NULL);
        end_time = tv.tv_sec * 1000000 + tv.tv_usec;
        execute_time = end_time - start_time;

        //okay, we got the result, print nano time and result
        printf("execute_time:%llu ##%s", execute_time, child_result);
    }
}