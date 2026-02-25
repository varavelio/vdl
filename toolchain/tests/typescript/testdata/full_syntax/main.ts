// Verifies full VDL syntax: constants, enums, patterns, types, and RPCs.

import type { Priority, Status } from "./gen/index.ts";
import {
  API_VERSION,
  isPriority,
  isStatus,
  MAX_PAGE_SIZE,
  type User,
  type UserServiceGetUserInput,
  type UserServiceGetUserOutput,
  type UserServiceUserActivityInput,
  type UserServiceUserActivityOutput,
  UserTopic,
  VDLPaths,
} from "./gen/index.ts";

function fail(name: string, expected: unknown, actual: unknown): never {
  console.error(`Verification failed for ${name}: expected "${expected}", got "${actual}"`);
  process.exit(1);
}

function main() {
  verifyConstants();
  verifyEnums();
  verifyPatterns();
  verifyTypes();
  verifyRPCs();
  console.log("Full syntax verification successful");
  process.exit(0);
}

function verifyConstants() {
  if (MAX_PAGE_SIZE !== 100) {
    fail("MAX_PAGE_SIZE", 100, MAX_PAGE_SIZE);
  }
  if (API_VERSION !== "1.0.0") {
    fail("API_VERSION", "1.0.0", API_VERSION);
  }
}

function verifyEnums() {
  // String enum
  const s: Status = "Active";
  if (s !== "Active") {
    fail("Status Active", "Active", s);
  }
  if (!isStatus(s)) {
    fail("isStatus('Active')", true, false);
  }

  // Int enum
  const p: Priority = 3;
  if (p !== 3) {
    fail("Priority High", 3, p);
  }
  if (!isPriority(p)) {
    fail("isPriority(3)", true, false);
  }
}

function verifyPatterns() {
  const topic = UserTopic("123", "login");
  const expected = "events.users.123.login";
  if (topic !== expected) {
    fail("UserTopic pattern", expected, topic);
  }
}

function verifyTypes() {
  // Verify struct fields and embedding/spreading
  const user: User = {
    id: "u1",
    createdAt: new Date(),
    updatedAt: new Date(),
    username: "alice",
    email: "alice@example.com",
    status: "Active",
    roles: ["admin"],
    preferences: { theme: "dark" },
    address: {
      street: "123 Main St",
      city: "Tech City",
      zip: "12345",
    },
  };

  // Verify the user object is valid
  if (user.username !== "alice") {
    fail("user.username", "alice", user.username);
  }

  // Verify optional field can be undefined
  const userWithBio: User = {
    ...user,
    bio: "Hello!",
  };
  if (userWithBio.bio !== "Hello!") {
    fail("user.bio", "Hello!", userWithBio.bio);
  }
}

function verifyRPCs() {
  // Verify RPC Catalog
  const path = VDLPaths.UserService.GetUser;
  const expectedPath = "/UserService/GetUser";
  if (path !== expectedPath) {
    fail("VDLPaths.UserService.GetUser", expectedPath, path);
  }

  // Verify Procedure Input/Output types compile
  const _input: UserServiceGetUserInput = { id: "1" };
  const _output: UserServiceGetUserOutput = {
    user: {
      id: "1",
      createdAt: new Date(),
      updatedAt: new Date(),
      username: "bob",
      email: "bob@example.com",
      status: "Active",
      roles: [],
      preferences: {},
      address: { street: "", city: "", zip: "" },
    },
  };

  // Verify Stream Input/Output types compile
  const _streamInput: UserServiceUserActivityInput = { userId: "1" };
  const _streamOutput: UserServiceUserActivityOutput = {
    action: "click",
    timestamp: new Date(),
  };

  // Suppress unused variable warnings
  void _input;
  void _output;
  void _streamInput;
  void _streamOutput;
}

main();
