const std = @import("std");
const builtin = @import("builtin");

pub fn build(b: *std.Build) void {
    // md4c needs libc; on linux default to musl so the binary is fully static.
    const target = b.standardTargetOptions(.{
        .default_target = .{ .abi = if (builtin.os.tag == .linux) .musl else null },
    });
    const optimize = b.standardOptimizeOption(.{});

    const md4c_dep = b.dependency("md4c", .{});
    const md4c_c = b.addTranslateC(.{
        .root_source_file = md4c_dep.path("src/md4c.h"),
        .target = target,
        .optimize = optimize,
    });

    const mod = b.createModule(.{
        .root_source_file = b.path("src/main.zig"),
        .target = target,
        .optimize = optimize,
        .link_libc = true,
        .strip = optimize != .debug,
        .single_threaded = true, // qwer never spawns threads; smaller binary
        .unwind_tables = .none,
        .imports = &.{.{ .name = "md4c", .module = md4c_c.createModule() }},
    });
    mod.addCSourceFile(.{ .file = md4c_dep.path("src/md4c.c") });

    const exe = b.addExecutable(.{ .name = "qwer", .root_module = mod });
    b.installArtifact(exe);

    const run = b.addRunArtifact(exe);
    run.addPassthruArgs();
    b.step("run", "Run qwer").dependOn(&run.step);

    const tests = b.addRunArtifact(b.addTest(.{ .root_module = mod }));
    b.step("test", "Run unit tests").dependOn(&tests.step);
}
