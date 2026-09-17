import { useEffect, useState } from "react";
import "./index.css";

const API = "http://localhost:8081";

function App() {
  const [polls, setPolls] = useState([]);
  const [selected, setSelected] = useState({});
  const [showCreate, setShowCreate] = useState(false);
  const [question, setQuestion] = useState("");
  const [options, setOptions] = useState(["", ""]);

  const loadPolls = async () => {
    try {
      const res = await fetch(`${API}/api/polls`);
      const data = await res.json();
      setPolls(data || []);
    } catch (error) {
      console.log("Backend connection error");
    }
  };

  useEffect(() => {
    loadPolls();
  }, []);

  const createPoll = async (e) => {
    e.preventDefault();

    const validOptions = options.filter((x) => x.trim());

    if (!question.trim() || validOptions.length < 2) {
      alert("Enter a question and at least 2 options");
      return;
    }

    try {
      const res = await fetch(`${API}/api/polls`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          question,
          options: validOptions,
        }),
      });

      if (!res.ok) {
        throw new Error("Failed to create poll");
      }

      setQuestion("");
      setOptions(["", ""]);
      setShowCreate(false);
      await loadPolls();

      alert("Poll created successfully!");
    } catch (error) {
      alert("Backend is not connected");
    }
  };

  const vote = (pollId, optionIndex) => {
    setSelected({
      ...selected,
      [pollId]: optionIndex,
    });
  };

  const submitVote = async (pollId) => {
    const optionIndex = selected[pollId];

    if (optionIndex === undefined) {
      alert("Please select an option");
      return;
    }

    try {
      const res = await fetch(`${API}/api/polls/${pollId}/vote`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          option: optionIndex,
        }),
      });

      if (!res.ok) {
        throw new Error("Vote failed");
      }

      await loadPolls();

      alert("Vote submitted successfully!");

      setSelected({
        ...selected,
        [pollId]: undefined,
      });
    } catch (error) {
      alert("Failed to submit vote");
    }
  };

  const getPercentage = (poll, index) => {
    const totalVotes = (poll.votes || []).reduce(
      (sum, vote) => sum + vote,
      0
    );

    if (totalVotes === 0) {
      return 0;
    }

    return Math.round(((poll.votes?.[index] || 0) / totalVotes) * 100);
  };

  return (
    <div className="app">
      <nav>
        <div className="logo">LivePoll</div>

        <div className="nav-links">
          <a href="#polls">Polls</a>
          <a href="#results">Results</a>

          <button onClick={() => setShowCreate(true)}>
            + Create Poll
          </button>
        </div>
      </nav>

      <section className="hero">
        <p className="tag">LIVE POLLING TOOL</p>

        <h1>
          Ask. Vote. <span>Decide.</span>
        </h1>

        <p>
          Create polls, collect votes and see what people think in real time.
        </p>

        <button
          className="hero-btn"
          onClick={() => setShowCreate(true)}
        >
          Create Your Poll
        </button>
      </section>

      <section className="poll-section" id="polls">
        <div className="section-title">
          <div>
            <p className="tag">ACTIVE POLLS</p>
            <h2>Cast Your Vote</h2>
          </div>

          <span>{polls.length} Polls</span>
        </div>

        {polls.length === 0 ? (
          <div className="empty">
            <h3>No polls yet</h3>
            <p>Create the first poll to get started.</p>
          </div>
        ) : (
          <div className="poll-grid">
            {polls.map((poll) => (
              <div className="poll-card" key={poll.id}>
                <h3>{poll.question}</h3>

                <div className="options">
                  {poll.options.map((option, index) => (
                    <button
                      key={index}
                      className={
                        selected[poll.id] === index
                          ? "option selected"
                          : "option"
                      }
                      onClick={() => vote(poll.id, index)}
                    >
                      <span className="radio">
                        {selected[poll.id] === index ? "✓" : ""}
                      </span>

                      {option}
                    </button>
                  ))}
                </div>

                <button
                  className="vote-btn"
                  disabled={selected[poll.id] === undefined}
                  onClick={() => submitVote(poll.id)}
                >
                  Submit Vote
                </button>
              </div>
            ))}
          </div>
        )}
      </section>

      <section className="results" id="results">
        <p className="tag">LIVE RESULTS</p>

        <h2>Poll Results</h2>

        {polls.length === 0 ? (
          <p>No results available yet.</p>
        ) : (
          polls.map((poll) => (
            <div className="result-card" key={poll.id}>
              <h3>{poll.question}</h3>

              {poll.options.map((option, index) => {
                const percentage = getPercentage(poll, index);
                const votes = poll.votes?.[index] || 0;

                return (
                  <div className="result-row" key={index}>
                    <div className="result-label">
                      <span>
                        {option} ({votes} vote{votes !== 1 ? "s" : ""})
                      </span>

                      <span>{percentage}%</span>
                    </div>

                    <div className="progress">
                      <div
                        className="progress-bar"
                        style={{
                          width: `${percentage}%`,
                        }}
                      ></div>
                    </div>
                  </div>
                );
              })}
            </div>
          ))
        )}
      </section>

      <footer>
        <strong>LivePoll</strong> — Ask. Vote. Decide.
      </footer>

      {showCreate && (
        <div className="modal">
          <div className="modal-box">
            <div className="modal-header">
              <h2>Create New Poll</h2>

              <button onClick={() => setShowCreate(false)}>
                ×
              </button>
            </div>

            <form onSubmit={createPoll}>
              <label>Poll Question</label>

              <input
                value={question}
                onChange={(e) => setQuestion(e.target.value)}
                placeholder="Enter your question..."
              />

              <label>Options</label>

              {options.map((option, index) => (
                <input
                  key={index}
                  value={option}
                  onChange={(e) => {
                    const copy = [...options];
                    copy[index] = e.target.value;
                    setOptions(copy);
                  }}
                  placeholder={`Option ${index + 1}`}
                />
              ))}

              <button
                type="button"
                className="add-btn"
                onClick={() => setOptions([...options, ""])}
              >
                + Add Option
              </button>

              <button className="create-btn" type="submit">
                Create Poll
              </button>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

export default App;