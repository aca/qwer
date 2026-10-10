const std = @import("std");
const Io = std.Io;
const Allocator = std.mem.Allocator;
const c = @import("md4c");

const Command = struct {
    level: c_uint = 0,
    code: []const u8 = "",
    children: std.StringArrayHashMapUnmanaged(*Command) = .empty,
};

/// Walks md4c events: "## heading" opens a command, the last code block
/// under it becomes its code. h1 headings are section titles only.
const Walker = struct {
    gpa: Allocator,
    root: *Command,
    stack: std.ArrayList(*Command) = .empty,
    last_heading: ?*Command = null,
    heading_level: c_uint = 0, // nonzero while inside a heading
    in_code: bool = false,
    buf: std.ArrayList(u8) = .empty,
    failed: bool = false,

    fn enterBlock(t: c.MD_BLOCKTYPE, detail: ?*anyopaque, ud: ?*anyopaque) callconv(.c) c_int {
        const w: *Walker = @ptrCast(@alignCast(ud));
        switch (t) {
            c.MD_BLOCK_H => {
                const d: *c.MD_BLOCK_H_DETAIL = @ptrCast(@alignCast(detail));
                w.heading_level = d.level;
                w.buf.clearRetainingCapacity();
            },
            c.MD_BLOCK_CODE => {
                w.in_code = true;
                w.buf.clearRetainingCapacity();
            },
            else => {},
        }
        return 0;
    }

    fn leaveBlock(t: c.MD_BLOCKTYPE, _: ?*anyopaque, ud: ?*anyopaque) callconv(.c) c_int {
        const w: *Walker = @ptrCast(@alignCast(ud));
        switch (t) {
            c.MD_BLOCK_H => {
                const level = w.heading_level;
                w.heading_level = 0;
                w.heading(level) catch {
                    w.failed = true;
                    return 1;
                };
            },
            c.MD_BLOCK_CODE => {
                w.in_code = false;
                if (w.last_heading) |h| {
                    h.code = w.gpa.dupe(u8, std.mem.trim(u8, w.buf.items, &std.ascii.whitespace)) catch {
                        w.failed = true;
                        return 1;
                    };
                }
            },
            else => {},
        }
        return 0;
    }

    fn span(_: c.MD_SPANTYPE, _: ?*anyopaque, _: ?*anyopaque) callconv(.c) c_int {
        return 0;
    }

    fn text(t: c.MD_TEXTTYPE, ptr: [*c]const c.MD_CHAR, size: c.MD_SIZE, ud: ?*anyopaque) callconv(.c) c_int {
        const w: *Walker = @ptrCast(@alignCast(ud));
        if (w.heading_level == 0 and !w.in_code) return 0;
        const s: []const u8 = switch (t) {
            c.MD_TEXT_NORMAL, c.MD_TEXT_CODE, c.MD_TEXT_ENTITY => ptr[0..size],
            c.MD_TEXT_NULLCHAR => "\u{FFFD}",
            else => return 0, // soft/hard breaks, raw html
        };
        w.buf.appendSlice(w.gpa, s) catch {
            w.failed = true;
            return 1;
        };
        return 0;
    }

    fn heading(w: *Walker, level: c_uint) !void {
        // A code block under an h1 must not attach to the previous command.
        if (level == 1) {
            w.last_heading = null;
            return;
        }
        const name = try w.gpa.dupe(u8, w.buf.items);
        for (name) |*ch| ch.* = if (ch.* == ' ') '-' else std.ascii.toLower(ch.*);

        const cmd = try w.gpa.create(Command);
        cmd.* = .{ .level = level };

        while (w.stack.items.len > 0 and w.stack.items[w.stack.items.len - 1].level >= level)
            _ = w.stack.pop();
        const parent = if (w.stack.items.len == 0) w.root else w.stack.items[w.stack.items.len - 1];
        try parent.children.put(w.gpa, name, cmd);

        try w.stack.append(w.gpa, cmd);
        w.last_heading = cmd;
    }
};

fn parseMarkdown(gpa: Allocator, input: []const u8) !*Command {
    const root = try gpa.create(Command);
    root.* = .{};
    var w: Walker = .{ .gpa = gpa, .root = root };
    const parser: c.MD_PARSER = .{
        .abi_version = 0,
        .flags = 0, // plain CommonMark
        .enter_block = Walker.enterBlock,
        .leave_block = Walker.leaveBlock,
        .enter_span = Walker.span, // md4c calls these unconditionally
        .leave_span = Walker.span,
        .text = Walker.text,
        .debug_log = null,
        .syntax = null,
    };
    const rc = c.md_parse(input.ptr, @intCast(input.len), &parser, &w);
    if (w.failed) return error.OutOfMemory;
    if (rc != 0) return error.ParseFailed;
    return root;
}

/// Children override parent code; new children are added.
fn mergeCommands(gpa: Allocator, parent: *Command, child: *Command) !void {
    var it = child.children.iterator();
    while (it.next()) |e| {
        if (parent.children.get(e.key_ptr.*)) |p| {
            if (e.value_ptr.*.code.len > 0) p.code = e.value_ptr.*.code;
            try mergeCommands(gpa, p, e.value_ptr.*);
        } else {
            try parent.children.put(gpa, e.key_ptr.*, e.value_ptr.*);
        }
    }
}

/// Every rune of `source` appears in `target` in order
/// (same as github.com/lithammer/fuzzysearch fuzzy.Find).
fn fuzzyMatch(source: []const u8, target: []const u8) bool {
    var t = target;
    var i: usize = 0;
    while (i < source.len) {
        const n = @min(std.unicode.utf8ByteSequenceLength(source[i]) catch 1, source.len - i);
        const pos = std.mem.indexOf(u8, t, source[i .. i + n]) orelse return false;
        t = t[pos + n ..];
        i += n;
    }
    return true;
}

fn findCommand(root: *Command, args: []const []const u8) ?*Command {
    var current = root;
    for (args) |arg| {
        if (current.children.get(arg)) |cmd| {
            current = cmd;
            continue;
        }
        // 0 or 2+ fuzzy matches: don't run anything
        var found: ?*Command = null;
        var it = current.children.iterator();
        while (it.next()) |e| {
            if (!fuzzyMatch(arg, e.key_ptr.*)) continue;
            if (found != null) return null;
            found = e.value_ptr.*;
        }
        current = found orelse return null;
    }
    return current;
}

fn listCommands(out: *Io.Writer, cmd: *Command, prefix: []const u8, gpa: Allocator) !void {
    var it = cmd.children.iterator();
    while (it.next()) |e| {
        const full = if (prefix.len == 0)
            e.key_ptr.*
        else
            try std.mem.concat(gpa, u8, &.{ prefix, " ", e.key_ptr.* });
        if (e.value_ptr.*.code.len > 0) try out.print("{s}\n", .{full});
        if (e.value_ptr.*.children.count() > 0) try listCommands(out, e.value_ptr.*, full, gpa);
    }
}

/// QWER.md (preferred) or qwer.md from cwd up to the git root (or "/"),
/// parent files first.
fn findAllCommandFiles(io: Io, gpa: Allocator) ![]const []const u8 {
    const cwd = Io.Dir.cwd();
    var files: std.ArrayList([]const u8) = .empty;
    var dir: []const u8 = try std.process.currentPathAlloc(io, gpa);
    while (true) {
        for ([_][]const u8{ "QWER.md", "qwer.md" }) |name| {
            const p = try std.fs.path.join(gpa, &.{ dir, name });
            if (cwd.access(io, p, .{})) |_| {
                try files.append(gpa, p);
                break;
            } else |_| {}
        }
        if (cwd.access(io, try std.fs.path.join(gpa, &.{ dir, ".git" }), .{})) |_| break else |_| {}
        dir = std.fs.path.dirname(dir) orelse break;
    }
    if (files.items.len == 0) return error.NoCommandFile;
    std.mem.reverse([]const u8, files.items);
    return files.items;
}

fn executeCommand(io: Io, gpa: Allocator, environ: *std.process.Environ.Map, out: *Io.Writer, code_in: []const u8, extra: []const []const u8) !std.process.Child.Term {
    const code = if (std.mem.startsWith(u8, code_in, "#!"))
        code_in
    else
        try std.mem.concat(gpa, u8, &.{ "#!/usr/bin/env bash\n", code_in });

    const tmpdir = environ.get("TMPDIR") orelse "/tmp";
    const cwd = Io.Dir.cwd();
    var path: []const u8 = undefined;
    var file: Io.File = undefined;
    while (true) {
        var r: [4]u8 = undefined;
        io.random(&r);
        path = try std.fmt.allocPrint(gpa, "{s}/qwer-{d}.sh", .{ tmpdir, std.mem.readInt(u32, &r, .little) });
        file = cwd.createFile(io, path, .{ .exclusive = true, .permissions = .fromMode(0o755) }) catch |err| switch (err) {
            error.PathAlreadyExists => continue,
            else => return err,
        };
        break;
    }
    defer cwd.deleteFile(io, path) catch {};
    {
        defer file.close(io);
        try file.writeStreamingAll(io, code);
    }

    try out.print("Running: {s}", .{path});
    if (extra.len > 0) {
        try out.writeAll(" ");
        for (extra, 0..) |a, i| try out.print("{s}{s}", .{ if (i > 0) " " else "", a });
    }
    try out.writeAll("\n");
    try out.flush();

    const argv = try std.mem.concat(gpa, []const u8, &.{ &.{path}, extra });
    var child = try std.process.spawn(io, .{ .argv = argv });
    return child.wait(io);
}

const usage_flags =
    \\  -f, --file string   Markdown file to read commands from
    \\  -h, --help          Show help
    \\      --list          List all available commands
    \\
;

const help_text =
    \\qwer - markdown-defined command runner (alternative to make, just, xc)
    \\
    \\qwer reads commands from a markdown file (QWER.md or qwer.md) and runs them.
    \\Each "## heading" defines a command; its fenced code block is the script.
    \\A "### heading" nested under it defines a subcommand. Scripts run with
    \\"#!/usr/bin/env bash" by default, or any interpreter via a shebang line.
    \\
    \\Usage:
    \\  qwer                                   List all available commands
    \\  qwer [command] [subcommand...]         Run a command
    \\  qwer [command] [args...]               Extra args are passed to the script ("$@")
    \\  qwer --list                            List all available commands
    \\  qwer --file FILE [command]             Use a specific markdown file
    \\
    \\Command file:
    \\  QWER.md is preferred over qwer.md. Files are searched from the current
    \\  directory up to the git repository root (or the filesystem root when not
    \\  in a git repo); commands in nested files override parent ones.
    \\
    \\  Example QWER.md:
    \\    ## build
    \\    ```
    \\    go build ./...
    \\    ```
    \\
    \\    ## deploy
    \\    ### prod
    \\    ```
    \\    #!/usr/bin/env python3
    \\    print("deploying prod")
    \\    ```
    \\
    \\  This defines commands "build" and "deploy prod".
    \\
    \\Command names are the heading text lowercased with spaces replaced by "-".
    \\Fuzzy matching is supported: "qwer bld" runs "build" if it is the only match.
    \\
    \\Flags:
    \\
;

const help_examples =
    \\
    \\Examples:
    \\  qwer build
    \\  qwer deploy prod
    \\  qwer test -- --verbose
    \\  qwer --file docs.md build
    \\
;

const Flags = struct {
    list: bool = false,
    help: bool = false,
    file: ?[]const u8 = null,
    positional: std.ArrayList([]const u8) = .empty,
};

fn parseBool(s: []const u8) ?bool {
    const yes = [_][]const u8{ "1", "t", "T", "TRUE", "true", "True" };
    const no = [_][]const u8{ "0", "f", "F", "FALSE", "false", "False" };
    for (yes) |y| if (std.mem.eql(u8, s, y)) return true;
    for (no) |n| if (std.mem.eql(u8, s, n)) return false;
    return null;
}

/// pflag-compatible parsing of --list, -h/--help, -f/--file (flags may be
/// interspersed with positional args). On error, writes pflag-style output
/// to `err_out` and returns error.Usage.
fn parseFlags(gpa: Allocator, argv: []const []const u8, err_out: *Io.Writer, argv0: []const u8) !Flags {
    var f: Flags = .{};
    var i: usize = 0;
    while (i < argv.len) : (i += 1) {
        const arg = argv[i];
        if (std.mem.startsWith(u8, arg, "--")) {
            const body = arg[2..];
            const eq = std.mem.indexOfScalar(u8, body, '=');
            const name = body[0 .. eq orelse body.len];
            const val: ?[]const u8 = if (eq) |e| body[e + 1 ..] else null;
            if (std.mem.eql(u8, name, "list") or std.mem.eql(u8, name, "help")) {
                const b = if (val) |v| parseBool(v) orelse
                    return usageError(err_out, argv0, "invalid argument \"{s}\" for \"--{s}\" flag: strconv.ParseBool: parsing \"{s}\": invalid syntax", .{ v, name, v }) else true;
                if (name[0] == 'l') f.list = b else f.help = b;
            } else if (std.mem.eql(u8, name, "file")) {
                if (val) |v| {
                    f.file = v;
                } else if (i + 1 < argv.len) {
                    i += 1;
                    f.file = argv[i];
                } else return usageError(err_out, argv0, "flag needs an argument: --file", .{});
            } else return usageError(err_out, argv0, "unknown flag: --{s}", .{name});
        } else if (arg.len > 1 and arg[0] == '-') {
            var j: usize = 1;
            while (j < arg.len) : (j += 1) switch (arg[j]) {
                'h' => f.help = true,
                'f' => {
                    if (j + 1 < arg.len) {
                        f.file = if (arg[j + 1] == '=') arg[j + 2 ..] else arg[j + 1 ..];
                    } else if (i + 1 < argv.len) {
                        i += 1;
                        f.file = argv[i];
                    } else return usageError(err_out, argv0, "flag needs an argument: 'f' in -f", .{});
                    break;
                },
                else => return usageError(err_out, argv0, "unknown shorthand flag: '{c}' in {s}", .{ arg[j], arg }),
            };
        } else {
            try f.positional.append(gpa, arg);
        }
    }
    return f;
}

fn usageError(out: *Io.Writer, argv0: []const u8, comptime fmt: []const u8, args: anytype) error{Usage} {
    out.print("Usage of {s}:\n" ++ usage_flags ++ fmt ++ "\n", .{argv0} ++ args) catch {};
    return error.Usage;
}

pub fn main(init: std.process.Init) !u8 {
    const io = init.io;
    const gpa = init.arena.allocator();

    var out_buf: [4096]u8 = undefined;
    var out_w: Io.File.Writer = .initStreaming(.stdout(), io, &out_buf);
    const out = &out_w.interface;
    defer out.flush() catch {};
    var err_buf: [1024]u8 = undefined;
    var err_w: Io.File.Writer = .initStreaming(.stderr(), io, &err_buf);
    const eout = &err_w.interface;
    defer eout.flush() catch {};

    const all_args = try init.minimal.args.toSlice(gpa);

    // Everything after "--" goes to the script verbatim.
    var argv: []const []const u8 = all_args[1..];
    var extra: []const []const u8 = &.{};
    for (argv, 0..) |a, i| if (std.mem.eql(u8, a, "--")) {
        extra = argv[i + 1 ..];
        argv = argv[0..i];
        break;
    };
    const flags = parseFlags(gpa, argv, eout, all_args[0]) catch |err| switch (err) {
        error.Usage => return 2,
        else => return err,
    };

    if (flags.help) {
        try out.writeAll(help_text);
        try out.flush();
        try eout.writeAll(usage_flags);
        try eout.flush();
        try out.writeAll(help_examples);
        return 0;
    }

    const files: []const []const u8 = if (flags.file) |f| &.{f} else findAllCommandFiles(io, gpa) catch |err| switch (err) {
        error.NoCommandFile => {
            try eout.writeAll("Error: no command file found (QWER.md or qwer.md) in current directory or parent directories\n");
            return 1;
        },
        else => return err,
    };

    const root = try gpa.create(Command);
    root.* = .{};
    for (files) |path| {
        const content = Io.Dir.cwd().readFileAlloc(io, path, gpa, .unlimited) catch |err| {
            try eout.print("Error reading {s}: {t}\n", .{ path, err });
            return 1;
        };
        try mergeCommands(gpa, root, try parseMarkdown(gpa, content));
    }

    const cmd_args = flags.positional.items;
    if (flags.list or cmd_args.len == 0) {
        try listCommands(out, root, "", gpa);
        return 0;
    }

    // Longest matching command path; the rest become script args.
    var n = cmd_args.len;
    const cmd = while (n > 0) : (n -= 1) {
        if (findCommand(root, cmd_args[0..n])) |found| break found;
    } else {
        try eout.print("Command not found: {s}\n", .{try std.mem.join(gpa, " ", cmd_args)});
        try eout.flush();
        try listCommands(out, root, "", gpa);
        return 1;
    };
    extra = try std.mem.concat(gpa, []const u8, &.{ cmd_args[n..], extra });

    const path = try std.mem.join(gpa, " ", cmd_args[0..n]);
    if (cmd.code.len == 0) {
        try eout.print("No code block found for command: {s}\n", .{path});
        try eout.flush();
        if (cmd.children.count() > 0) try listCommands(out, cmd, path, gpa);
        return 1;
    }

    const term = executeCommand(io, gpa, init.environ_map, out, cmd.code, extra) catch |err| {
        try eout.print("Command failed: {t}\n", .{err});
        return 1;
    };
    switch (term) {
        .exited => |status| if (status == 0) return 0 else try eout.print("Command failed: exit status {d}\n", .{status}),
        .signal => |sig| try eout.print("Command failed: signal: {t}\n", .{sig}),
        else => try eout.print("Command failed: {any}\n", .{term}),
    }
    return 1;
}

test "parse headings and code blocks" {
    var arena: std.heap.ArenaAllocator = .init(std.testing.allocator);
    defer arena.deinit();
    const md =
        \\# Title
        \\```
        \\ignored
        \\```
        \\## Echo Two
        \\```sh
        \\echo hi
        \\```
        \\### sub
        \\    indented code
        \\## `Code` *span*
        \\```
        \\x
        \\```
        \\## other
        \\> ```
        \\> quoted
        \\> ```
    ;
    const root = try parseMarkdown(arena.allocator(), md);
    const echo = root.children.get("echo-two").?;
    try std.testing.expectEqualStrings("echo hi", echo.code);
    try std.testing.expectEqualStrings("indented code", echo.children.get("sub").?.code);
    try std.testing.expectEqualStrings("quoted", root.children.get("other").?.code);
    try std.testing.expectEqualStrings("x", root.children.get("code-span").?.code);
    try std.testing.expectEqual(3, root.children.count());
}

test fuzzyMatch {
    try std.testing.expect(fuzzyMatch("e2", "echo2"));
    try std.testing.expect(!fuzzyMatch("2e", "echo2"));
    try std.testing.expect(fuzzyMatch("", "x"));
    try std.testing.expect(fuzzyMatch("한", "한국"));
}
