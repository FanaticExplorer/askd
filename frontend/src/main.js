import { Browser, Events } from "@wailsio/runtime";
import { AppState } from "../bindings/github.com/FanaticExplorer/askd";

document.getElementById("empty-state").addEventListener("click", (e) => {
    const a = e.target.closest("a");
    if (a) {
        e.preventDefault();
        Browser.OpenURL(a.href);
    }
});

let currentSession = null;
let currentIndex = 0;
let answers = [];

const emptyState = document.getElementById("empty-state");
const questionsEl = document.getElementById("questions");
const progressFill = document.getElementById("progress-fill");
const progressText = document.getElementById("progress-text");
const titleEl = document.getElementById("q-title");
const descEl = document.getElementById("q-desc");
const optionsEl = document.getElementById("q-options");
const customEl = document.getElementById("q-custom");
const customInput = document.getElementById("custom-input");
const validationMsg = document.getElementById("validation-msg");
const btnBack = document.getElementById("btn-back");
const btnNext = document.getElementById("btn-next");
const btnSubmit = document.getElementById("btn-submit");

Events.On("new-session", (event) => {
    currentSession = event.data;
    currentIndex = 0;
    answers = currentSession.questions.map(() => ({
        selectedIndexes: [],
        customText: "",
    }));
    clearValidation();
    showQuestions();
    renderQuestion();
});

Events.On("session-timeout", () => {
    showEmpty();
});

function showQuestions() {
    emptyState.style.display = "none";
    questionsEl.style.display = "block";
}

function showEmpty() {
    emptyState.style.display = "flex";
    questionsEl.style.display = "none";
    currentSession = null;
}

function updateProgress() {
    const total = currentSession.questions.length;
    const pct = total > 0 ? ((currentIndex + 1) / total) * 100 : 0;
    progressFill.style.width = pct + "%";
    progressText.textContent = "Question " + (currentIndex + 1) + " of " + total;
}

function clearValidation() {
    validationMsg.textContent = "";
    validationMsg.classList.remove("visible");
}

function showValidation(msg) {
    validationMsg.textContent = msg;
    validationMsg.classList.add("visible");
}

function renderQuestion() {
    const q = currentSession.questions[currentIndex];

    updateProgress();
    titleEl.textContent = q.title;
    descEl.textContent = q.description || "";
    clearValidation();

    optionsEl.innerHTML = "";
    const ans = answers[currentIndex];
    const inputType = q.multiSelect ? "checkbox" : "radio";
    const groupName = "q-opt-" + currentIndex;

    q.options.forEach((opt, i) => {
        const row = document.createElement("div");
        row.className = "option-row";

        const input = document.createElement("input");
        input.type = inputType;
        input.name = groupName;
        input.value = i;
        if (ans.selectedIndexes.includes(i)) {
            input.checked = true;
        }

        input.addEventListener("change", () => {
            clearValidation();
            if (q.multiSelect) {
                if (input.checked) {
                    ans.selectedIndexes.push(i);
                } else {
                    ans.selectedIndexes = ans.selectedIndexes.filter(
                        (idx) => idx !== i
                    );
                }
            } else {
                ans.selectedIndexes = [i];
            }
        });

        const textBlock = document.createElement("span");
        textBlock.className = "option-text";

        const labelSpan = document.createElement("span");
        labelSpan.className = "option-label";
        labelSpan.textContent = opt.label;

        if (opt.recommended) {
            const rec = document.createElement("span");
            rec.className = "rec-pill";
            rec.textContent = "recommended";
            labelSpan.appendChild(rec);
        }

        textBlock.appendChild(labelSpan);

        if (opt.description) {
            const desc = document.createElement("span");
            desc.className = "option-desc";
            desc.textContent = opt.description;
            textBlock.appendChild(desc);
        }

        row.appendChild(input);
        row.appendChild(textBlock);

        row.addEventListener("click", (e) => {
            if (e.target === input) return;
            input.checked = !input.checked;
            input.dispatchEvent(new Event("change", { bubbles: true }));
        });

        optionsEl.appendChild(row);
    });

    if (q.allowCustom) {
        customEl.style.display = "block";
        customInput.value = ans.customText;
    } else {
        customEl.style.display = "none";
    }

    btnBack.style.display = currentIndex > 0 ? "" : "none";
    if (currentIndex < currentSession.questions.length - 1) {
        btnNext.style.display = "";
        btnSubmit.style.display = "none";
    } else {
        btnNext.style.display = "none";
        btnSubmit.style.display = "";
    }
}

function saveCurrentAnswer() {
    const ans = answers[currentIndex];
    if (currentSession.questions[currentIndex].allowCustom) {
        ans.customText = customInput.value;
    }
}

btnBack.addEventListener("click", () => {
    saveCurrentAnswer();
    currentIndex--;
    renderQuestion();
});

btnNext.addEventListener("click", () => {
    saveCurrentAnswer();
    currentIndex++;
    renderQuestion();
});

btnSubmit.addEventListener("click", () => {
    saveCurrentAnswer();

    for (let i = 0; i < answers.length; i++) {
        const q = currentSession.questions[i];
        const a = answers[i];
        const hasSelection = a.selectedIndexes.length > 0;
        const hasCustom = q.allowCustom && a.customText.trim() !== "";
        if (!hasSelection && !hasCustom) {
            currentIndex = i;
            renderQuestion();
            showValidation("Answer this question to continue: " + q.title);
            return;
        }
    }

    AppState.SubmitAnswers({ answers: [...answers] });
    showEmpty();
});
