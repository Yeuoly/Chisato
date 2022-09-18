#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <sys/time.h>
#include <sys/wait.h>
#include <sys/types.h>
#include <sys/user.h>
#include <sys/stat.h>
#include <sys/syscall.h>
#include <fcntl.h>
#include <string.h>
#include <errno.h>
#include <signal.h>
#include <sys/ptrace.h>

#define ull unsigned long long int
#define cptr char *

#define max(a,b) (a>b?a:b)

#define MAX_TIME 5
#define MAX_MEMORY 128*1024*1024

ull execute_time = 0;
ull execute_memory = 0;
cptr child_result = NULL;
cptr child_err = NULL;
ull child_result_size = 0;
ull child_err_size = 0;
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

void appendError(cptr to_append, ull size) {
    if (child_err == NULL) {
        child_err = malloc(size);
        child_err_size = size;
    } else {
        child_err = realloc(child_err, child_err_size + size);
        child_err_size += size;
    }
    for (ull i = 0; i < size; i++) {
        child_err[child_err_size - size + i] = to_append[i];
    }
}

ull get_process_memory(pid_t pid) {
    ull vmData = 0;
    ull vmStk = 0;
    char path[64] = { 0 };
    sprintf(path, "/proc/%d/status", pid);
    int fd = open(path, O_RDONLY);
    char buf[4096] = { 0 };
    read(fd, buf, sizeof(buf));
    close(fd);
    char *p = strstr(buf, "VmData:");
    p += strlen("VmData:");
    while (*p == ' ' || *p == '\t') p++;
    while (*p >= '0' && *p <= '9') {
        vmData = vmData * 10 + *p - '0';
        p++;
    }
    //VmStk
    p = strstr(buf, "VmStk:");
    p += strlen("VmStk:");
    while (*p == ' ' || *p == '\t') p++;
    while (*p >= '0' && *p <= '9') {
        vmStk = vmStk * 10 + *p - '0';
        p++;
    }
    //data is kB, convert to bytes
    return (vmData + vmStk) * 1024;
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
        alarm(MAX_TIME);
        //child process
        //close unused pipe
        close(pipe_stdout_read);
        close(pipe_stdin_write);

        //redirect stdout
        //stdin will be read to child process, but stdout will be hijack
        dup2(pipe_stdout_write, STDOUT_FILENO);

        ptrace(PTRACE_TRACEME, NULL, NULL);
        raise(SIGSTOP);
        //execute command
        execvp(args[1], args + 1);
    } else {
        //parent process
        //close unused pipe
        close(pipe_stdout_write);
        close(pipe_stdin_read);

        int status;
        // Wait for child to stop itself.
        waitpid(pid, &status, 0);
        ptrace(PTRACE_SETOPTIONS, pid, NULL, PTRACE_O_TRACESYSGOOD);
        //get current microsecond
        struct timeval tv;
        gettimeofday(&tv, NULL);
        start_time = tv.tv_sec * 1000000 + tv.tv_usec;
        while(1) {
            //capture syscall
            ptrace(PTRACE_SYSCALL, pid, NULL, NULL);
            waitpid(pid, &status, 0);
            if (WIFEXITED(status)) {
                break;
            }
            //check signal for alarm while dead loop
            if (WIFSTOPPED(status)) {
                //get stop signal
                int sig = WSTOPSIG(status);
                if (sig == SIGALRM) {
                    //kill child process
                    appendError("Time Limit Exceeded ##", 21);
                    kill(pid, SIGKILL);
                    break;
                }
            }

            if (WIFSIGNALED(status)) {
                int sig = WTERMSIG(status);
                if (sig == SIGALRM) {
                    appendError("Time Limit Exceeded ##", 21);
                    kill(pid, SIGKILL);
                    break;
                }
                if (sig == SIGKILL) {
                    appendError("Process has been killed by Chisato ##", 36);
                    break;
                }
            }
            struct user_regs_struct regs;
            ptrace(PTRACE_GETREGS, pid, NULL, &regs);
            //get syscall number
            int syscall = regs.orig_rax;
            //block network
            if(syscall == SYS_socket || syscall == SYS_connect || syscall == SYS_accept || syscall == SYS_bind || syscall == SYS_listen) {
                regs.rax = -1;
                ptrace(PTRACE_SETREGS, pid, NULL, &regs);
                kill(pid, SIGKILL);
            }
            //block create link unlink ...
            if(syscall == SYS_link || syscall == SYS_unlink || syscall == SYS_symlink || syscall == SYS_rename || syscall == SYS_mkdir || syscall == SYS_rmdir) {
                regs.rax = -1;
                ptrace(PTRACE_SETREGS, pid, NULL, &regs);
                kill(pid, SIGKILL);
            }
            //block chown chmod ...
            if(syscall == SYS_chown || syscall == SYS_fchown || syscall == SYS_lchown || syscall == SYS_fchmod || syscall == SYS_chmod) {
                regs.rax = -1;
                ptrace(PTRACE_SETREGS, pid, NULL, &regs);
                kill(pid, SIGKILL);
            }
            //block ls
            if(syscall == SYS_getdents || syscall == SYS_getdents64) {
                regs.rax = -1;
                ptrace(PTRACE_SETREGS, pid, NULL, &regs);
                kill(pid, SIGKILL);
            }
            //block getuid setuid and so on
            if(syscall == SYS_getuid || syscall == SYS_setuid || syscall == SYS_geteuid || syscall == SYS_getgid || syscall == SYS_setgid || syscall == SYS_getegid) {
                regs.rax = -1;
                ptrace(PTRACE_SETREGS, pid, NULL, &regs);
                kill(pid, SIGKILL);
            }
            //block ptrace, fork, vfork
            if(syscall == SYS_ptrace || syscall == SYS_fork || syscall == SYS_vfork) {
                regs.rax = -1;
                ptrace(PTRACE_SETREGS, pid, NULL, &regs);
                kill(pid, SIGKILL);
            }
            //block linux page scheduling
            if(syscall == SYS_madvise || syscall == SYS_mlock || syscall == SYS_mlockall || syscall == SYS_munlock || syscall == SYS_munlockall) {
                regs.rax = -1;
                ptrace(PTRACE_SETREGS, pid, NULL, &regs);
                kill(pid, SIGKILL);
            }

            //get output
            if(syscall == SYS_write) {
                //child process has been blocked
                ptrace(PTRACE_SYSCALL, pid, NULL, NULL);
                if (regs.rdi == 1 || regs.rdi == 2) {
                    //read from stdout
                    int length = regs.rdx;
                    char *buf = malloc(length);
                    int read_length = 0;
                    while (read_length < length) {
                        int n = read(pipe_stdout_read, buf + read_length, length - read_length);
                        if (n == -1) {
                            break;
                        }
                        read_length += n;
                    }
                    append(buf, length);
                }
            }
            //capture syscall which will allocate memory
            if(syscall == SYS_brk || syscall == SYS_mmap || syscall == SYS_mremap || syscall == SYS_munmap || syscall == SYS_mprotect) {
                //get current memory usage
                execute_memory = max(execute_memory, get_process_memory(pid));
                if (execute_memory > MAX_MEMORY) {
                    kill(pid, SIGKILL);
                    break;
                }
            }
        }

        //process exit
        waitpid(pid, &status, 0);
        //get nano time
        gettimeofday(&tv, NULL);
        end_time = tv.tv_sec * 1000000 + tv.tv_usec;
        execute_time = end_time - start_time;

        //okay, we got the result
        struct header {
            ull time;
            ull memory;
            ull length;
            ull has_error;
        };
        struct header header;
        header.time = execute_time;
        header.memory = execute_memory;

        //check if we got error
        if (child_err_size > 0) {
            header.has_error = 1;
            header.length = child_err_size;
            write(STDOUT_FILENO, &header, sizeof(header));
            write(STDOUT_FILENO, child_err, child_err_size);
        } else {
            header.has_error = 0;
            header.length = child_result_size;
            write(STDOUT_FILENO, &header, sizeof(header));
            write(STDOUT_FILENO, child_result, child_result_size);
        }
    }
}