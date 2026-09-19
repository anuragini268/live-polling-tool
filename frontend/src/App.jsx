import { useEffect, useState } from "react";

const API = "http://localhost:8081/api";

function App() {
  const [page, setPage] = useState("login");
  const [token, setToken] = useState(localStorage.getItem("token") || "");
  const [user, setUser] = useState(
    JSON.parse(localStorage.getItem("user") || "null")
  );

  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  const [question, setQuestion] = useState("");
  const [options, setOptions] = useState(["", "", "", ""]);

  const [poll, setPoll] = useState(null);
  const [selectedOption, setSelectedOption] = useState(null);
  const [message, setMessage] = useState("");
  const [loading, setLoading] = useState(false);

  const path = window.location.pathname;
  const pollId = path.startsWith("/poll/")
    ? path.split("/poll/")[1]
    : null;

  // Generate browser-specific voter ID
  const getVoterId = () => {
    let id = localStorage.getItem("voter_id");

    if (!id) {
      id = "voter-" + crypto.randomUUID();
      localStorage.setItem("voter_id", id);
    }

    return id;
  };

  // Load public poll
  useEffect(() => {
    if (!pollId) return;

    setPage("vote");

    fetch(`${API}/polls/${pollId}`)
      .then(async (res) => {
        const data = await res.json();

        if (!res.ok) {
          throw new Error(data.error || "Poll not found");
        }

        setPoll(data);
      })
      .catch((err) => {
        setMessage(err.message);
      });
  }, [pollId]);

  // SSE live updates
  useEffect(() => {
    if (!pollId) return;

    const eventSource = new EventSource(
      `${API}/polls/${pollId}/stream`
    );

    eventSource.addEventListener("poll", (event) => {
      try {
        const updatedPoll = JSON.parse(event.data);
        setPoll(updatedPoll);
      } catch (error) {
        console.error("SSE parse error:", error);
      }
    });

    eventSource.onerror = () => {
      console.log("SSE connection temporarily unavailable");
    };

    return () => {
      eventSource.close();
    };
  }, [pollId]);

  // SIGNUP
  const handleSignup = async (e) => {
    e.preventDefault();
    setLoading(true);
    setMessage("");

    try {
      const response = await fetch(`${API}/auth/signup`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          name: name.trim(),
          email: email.trim().toLowerCase(),
          password: password,
        }),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(
          data.error || "Failed to create account"
        );
      }

      localStorage.setItem("token", data.token);
      localStorage.setItem(
        "user",
        JSON.stringify(data.user)
      );

      setToken(data.token);
      setUser(data.user);
      setPage("dashboard");
      setMessage("Account created successfully!");

      setName("");
      setEmail("");
      setPassword("");
    } catch (error) {
      console.error("Signup error:", error);
      setMessage(error.message);
    } finally {
      setLoading(false);
    }
  };

  // LOGIN
  const handleLogin = async (e) => {
    e.preventDefault();
    setLoading(true);
    setMessage("");

    try {
      const res = await fetch(`${API}/auth/login`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          email: email.trim().toLowerCase(),
          password,
        }),
      });

      const data = await res.json();

      if (!res.ok) {
        throw new Error(data.error || "Login failed");
      }

      localStorage.setItem("token", data.token);
      localStorage.setItem(
        "user",
        JSON.stringify(data.user)
      );

      setToken(data.token);
      setUser(data.user);
      setPage("dashboard");
      setMessage("Login successful!");

      setEmail("");
      setPassword("");
    } catch (error) {
      console.error("Login error:", error);
      setMessage(error.message);
    } finally {
      setLoading(false);
    }
  };

  // CREATE POLL
  const handleCreatePoll = async (e) => {
    e.preventDefault();
    setLoading(true);
    setMessage("");

    const cleanedOptions = options
      .map((option) => option.trim())
      .filter(Boolean);

    if (!question.trim()) {
      setMessage("Please enter a question.");
      setLoading(false);
      return;
    }

    if (cleanedOptions.length < 2) {
      setMessage("Please enter at least 2 options.");
      setLoading(false);
      return;
    }

    try {
      const res = await fetch(`${API}/polls`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          question: question.trim(),
          options: cleanedOptions,
        }),
      });

      const data = await res.json();

      if (!res.ok) {
        throw new Error(
          data.error || "Failed to create poll"
        );
      }

      setPoll(data);
      setQuestion("");
      setOptions(["", "", "", ""]);
      setMessage("Poll created successfully!");
    } catch (error) {
      console.error("Create poll error:", error);
      setMessage(error.message);
    } finally {
      setLoading(false);
    }
  };

  // VOTE
  const handleVote = async () => {
    if (selectedOption === null) {
      setMessage("Please select an option.");
      return;
    }

    if (!poll) return;

    setLoading(true);
    setMessage("");

    try {
      const res = await fetch(
        `${API}/polls/${poll.id}/vote`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            option: selectedOption,
            voter_id: getVoterId(),
          }),
        }
      );

      const data = await res.json();

      if (!res.ok) {
        throw new Error(data.error || "Vote failed");
      }

      setMessage("Your vote has been recorded!");
    } catch (error) {
      console.error("Vote error:", error);
      setMessage(error.message);
    } finally {
      setLoading(false);
    }
  };

  // LOGOUT
  const logout = () => {
    localStorage.removeItem("token");
    localStorage.removeItem("user");

    setToken("");
    setUser(null);
    setPage("login");
    setPoll(null);
    setMessage("Logged out successfully.");
  };

  // COPY SHARE LINK
  const copyShareLink = async () => {
    if (!poll) return;

    const link =
      `${window.location.origin}/poll/${poll.id}`;

    try {
      await navigator.clipboard.writeText(link);
      setMessage("Share link copied!");
    } catch (error) {
      setMessage("Unable to copy link.");
    }
  };

  // UPDATE OPTION
  const updateOption = (index, value) => {
    const updated = [...options];
    updated[index] = value;
    setOptions(updated);
  };

  // TOTAL VOTES
  const totalVotes =
    poll?.votes?.reduce(
      (sum, value) => sum + value,
      0
    ) || 0;

  // PERCENTAGE
  const percentage = (votes) => {
    if (!totalVotes) return 0;

    return Math.round(
      (votes / totalVotes) * 100
    );
  };

  // =========================
  // PUBLIC VOTING PAGE
  // =========================
  if (page === "vote") {
    return (
      <div className="app">
        <div className="card poll-card">
          <h1>Live Poll</h1>

          {message && (
            <div className="message">
              {message}
            </div>
          )}

          {!poll ? (
            <p>Loading poll...</p>
          ) : (
            <>
              <h2>{poll.question}</h2>

              <div className="vote-options">
                {poll.options.map(
                  (option, index) => (
                    <label
                      className={
                        selectedOption === index
                          ? "vote-option selected"
                          : "vote-option"
                      }
                      key={index}
                    >
                      <input
                        type="radio"
                        name="poll-option"
                        checked={
                          selectedOption === index
                        }
                        onChange={() =>
                          setSelectedOption(index)
                        }
                      />

                      <span>{option}</span>
                    </label>
                  )
                )}
              </div>

              <button
                className="primary-btn"
                onClick={handleVote}
                disabled={loading}
              >
                {loading
                  ? "Voting..."
                  : "Submit Vote"}
              </button>

              <div className="results">
                <h3>Live Results</h3>

                {poll.options.map(
                  (option, index) => (
                    <div
                      className="result-row"
                      key={index}
                    >
                      <div className="result-header">
                        <span>{option}</span>

                        <span>
                          {poll.votes[index] || 0}{" "}
                          votes (
                          {percentage(
                            poll.votes[index] || 0
                          )}
                          %)
                        </span>
                      </div>

                      <div className="bar">
                        <div
                          className="bar-fill"
                          style={{
                            width: `${percentage(
                              poll.votes[index] || 0
                            )}%`,
                          }}
                        />
                      </div>
                    </div>
                  )
                )}

                <p className="total">
                  Total votes: {totalVotes}
                </p>
              </div>
            </>
          )}
        </div>
      </div>
    );
  }

  // =========================
  // LOGIN PAGE
  // =========================
  if (!token && page === "login") {
    return (
      <div className="app">
        <div className="card auth-card">
          <h1>Live Polling Tool</h1>

          <p className="subtitle">
            Login to create and manage polls
          </p>

          {message && (
            <div className="message">
              {message}
            </div>
          )}

          <form onSubmit={handleLogin}>
            <input
              type="email"
              placeholder="Email"
              value={email}
              onChange={(e) =>
                setEmail(e.target.value)
              }
              required
            />

            <input
              type="password"
              placeholder="Password"
              value={password}
              onChange={(e) =>
                setPassword(e.target.value)
              }
              required
            />

            <button
              className="primary-btn"
              disabled={loading}
            >
              {loading
                ? "Logging in..."
                : "Login"}
            </button>
          </form>

          <p>
            Don't have an account?{" "}
            <button
              type="button"
              className="link-btn"
              onClick={() => {
                setPage("signup");
                setMessage("");
              }}
            >
              Sign up
            </button>
          </p>
        </div>
      </div>
    );
  }

  // =========================
  // SIGNUP PAGE
  // =========================
  if (!token && page === "signup") {
    return (
      <div className="app">
        <div className="card auth-card">
          <h1>Create Account</h1>

          <p className="subtitle">
            Start creating live polls
          </p>

          {message && (
            <div className="message">
              {message}
            </div>
          )}

          <form onSubmit={handleSignup}>
            <input
              type="text"
              placeholder="Full Name"
              value={name}
              onChange={(e) =>
                setName(e.target.value)
              }
              required
            />

            <input
              type="email"
              placeholder="Email"
              value={email}
              onChange={(e) =>
                setEmail(e.target.value)
              }
              required
            />

            <input
              type="password"
              placeholder="Password (min 6 characters)"
              value={password}
              onChange={(e) =>
                setPassword(e.target.value)
              }
              minLength={6}
              required
            />

            <button
              className="primary-btn"
              disabled={loading}
            >
              {loading
                ? "Creating..."
                : "Create Account"}
            </button>
          </form>

          <p>
            Already have an account?{" "}
            <button
              type="button"
              className="link-btn"
              onClick={() => {
                setPage("login");
                setMessage("");
              }}
            >
              Login
            </button>
          </p>
        </div>
      </div>
    );
  }

  // =========================
  // DASHBOARD
  // =========================
  return (
    <div className="app">
      <header className="navbar">
        <div>
          <h2>Live Polling Tool</h2>

          <span>
            Welcome, {user?.name || "User"}
          </span>
        </div>

        <button
          className="logout-btn"
          onClick={logout}
        >
          Logout
        </button>
      </header>

      <main className="dashboard">
        <div className="card">
          <h1>Create a Poll</h1>

          <p className="subtitle">
            Create a poll and share it with your
            audience.
          </p>

          {message && (
            <div className="message">
              {message}
            </div>
          )}

          <form onSubmit={handleCreatePoll}>
            <input
              type="text"
              placeholder="Enter your question"
              value={question}
              onChange={(e) =>
                setQuestion(e.target.value)
              }
              required
            />

            {options.map(
              (option, index) => (
                <input
                  key={index}
                  type="text"
                  placeholder={`Option ${index + 1}`}
                  value={option}
                  onChange={(e) =>
                    updateOption(
                      index,
                      e.target.value
                    )
                  }
                />
              )
            )}

            <button
              className="primary-btn"
              disabled={loading}
            >
              {loading
                ? "Creating..."
                : "Create Poll"}
            </button>
          </form>
        </div>

        {poll && (
          <div className="card">
            <h2>Your Poll</h2>

            <h3>{poll.question}</h3>

            <div className="share-box">
              <input
                value={`${window.location.origin}/poll/${poll.id}`}
                readOnly
              />

              <button
                type="button"
                className="secondary-btn"
                onClick={copyShareLink}
              >
                Copy Link
              </button>
            </div>

            <div className="results">
              <h3>Live Results</h3>

              {poll.options.map(
                (option, index) => (
                  <div
                    className="result-row"
                    key={index}
                  >
                    <div className="result-header">
                      <span>{option}</span>

                      <span>
                        {poll.votes[index] || 0}{" "}
                        votes (
                        {percentage(
                          poll.votes[index] || 0
                        )}
                        %)
                      </span>
                    </div>

                    <div className="bar">
                      <div
                        className="bar-fill"
                        style={{
                          width: `${percentage(
                            poll.votes[index] || 0
                          )}%`,
                        }}
                      />
                    </div>
                  </div>
                )
              )}

              <p className="total">
                Total votes: {totalVotes}
              </p>
            </div>
          </div>
        )}
      </main>
    </div>
  );
}

export default App;