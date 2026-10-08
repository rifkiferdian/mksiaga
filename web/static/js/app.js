document.querySelectorAll("[data-password-toggle]").forEach((button) => {
  button.addEventListener("click", () => {
    const input = document.getElementById(button.dataset.target);
    if (!input) return;

    const visible = input.type === "text";
    input.type = visible ? "password" : "text";
    button.setAttribute("aria-pressed", String(!visible));
    button.setAttribute("aria-label", visible ? "Tampilkan password" : "Sembunyikan password");
    button.querySelector("[data-eye-open]")?.classList.toggle("hidden", !visible);
    button.querySelector("[data-eye-closed]")?.classList.toggle("hidden", visible);
  });
});

const dashboardSidebar = document.getElementById("dashboardSidebar");
const sidebarBackdrop = document.getElementById("sidebarBackdrop");

function setSidebar(open) {
  if (!dashboardSidebar || !sidebarBackdrop) return;
  dashboardSidebar.classList.toggle("-translate-x-full", !open);
  sidebarBackdrop.classList.toggle("hidden", !open);
  document.body.classList.toggle("overflow-hidden", open && window.innerWidth < 1024);
}

document.getElementById("openSidebar")?.addEventListener("click", () => setSidebar(true));
document.getElementById("closeSidebar")?.addEventListener("click", () => setSidebar(false));
sidebarBackdrop?.addEventListener("click", () => setSidebar(false));

const toast = document.getElementById("dashboardToast");
let toastTimer;

function showDashboardToast(message) {
  if (!toast) return;
  toast.textContent = message;
  toast.classList.remove("hidden");
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => toast.classList.add("hidden"), 2600);
}

document.querySelectorAll("[data-demo-action]").forEach((button) => {
  button.addEventListener("click", () => showDashboardToast(button.dataset.demoAction));
});

const recordRows = [...document.querySelectorAll("[data-record]")];
const recordSearch = document.getElementById("recordSearch");
const recordCount = document.getElementById("recordCount");
const recordEmpty = document.getElementById("recordEmpty");
let activeRecordFilter = "Semua";

function filterRecords() {
  const query = recordSearch?.value.trim().toLowerCase() ?? "";
  let visible = 0;
  recordRows.forEach((row) => {
    const matchesFilter = activeRecordFilter === "Semua" || row.dataset.status === activeRecordFilter;
    const matchesSearch = (row.dataset.search ?? "").includes(query);
    const shown = matchesFilter && matchesSearch;
    row.classList.toggle("hidden", !shown);
    if (shown) visible += 1;
  });
  if (recordCount) recordCount.textContent = `${visible} catatan`;
  recordEmpty?.classList.toggle("hidden", visible !== 0);
}

recordSearch?.addEventListener("input", filterRecords);
document.querySelectorAll("[data-record-filter]").forEach((button) => {
  button.addEventListener("click", () => {
    activeRecordFilter = button.dataset.recordFilter;
    document.querySelectorAll("[data-record-filter]").forEach((item) => {
      const active = item === button;
      item.classList.toggle("bg-orange-50", active);
      item.classList.toggle("font-bold", active);
      item.classList.toggle("text-orange-700", active);
      item.classList.toggle("text-slate-500", !active);
    });
    filterRecords();
  });
});

const recordModal = document.getElementById("recordModal");

function setText(id, value) {
  const element = document.getElementById(id);
  if (element) element.textContent = value;
}

function openRecordModal(row) {
  if (!recordModal) return;
  const initials = (row.dataset.name ?? "").split(/\s+/).map((part) => part[0]).slice(0, 2).join("");
  setText("recordModalTitle", row.dataset.name);
  setText("recordModalInitials", initials);
  setText("recordModalAlias", row.dataset.alias);
  setText("recordModalStatus", row.dataset.status);
  setText("recordModalLocation", row.dataset.location);
  setText("recordModalDate", row.dataset.date);
  setText("recordModalNote", row.dataset.note);
  recordModal.classList.remove("hidden");
  recordModal.classList.add("flex");
  document.body.classList.add("overflow-hidden");
}

function closeRecordModal() {
  if (!recordModal) return;
  recordModal.classList.add("hidden");
  recordModal.classList.remove("flex");
  document.body.classList.remove("overflow-hidden");
}

recordRows.forEach((row) => {
  row.addEventListener("click", () => openRecordModal(row));
  row.addEventListener("keydown", (event) => {
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      openRecordModal(row);
    }
  });
});
document.getElementById("closeRecordModal")?.addEventListener("click", closeRecordModal);
recordModal?.addEventListener("click", (event) => {
  if (event.target === recordModal) closeRecordModal();
});
document.addEventListener("keydown", (event) => {
  if (event.key === "Escape") {
    closeRecordModal();
    setSidebar(false);
  }
});
