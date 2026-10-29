// randomUUID hanya tersedia pada secure context di sebagian browser.
// Sediakan fallback agar fitur frontend tetap berjalan pada server HTTP lama.
if (globalThis.crypto && typeof globalThis.crypto.randomUUID !== "function") {
  globalThis.crypto.randomUUID = () => {
    const bytes = new Uint8Array(16);
    globalThis.crypto.getRandomValues(bytes);
    bytes[6] = (bytes[6] & 0x0f) | 0x40;
    bytes[8] = (bytes[8] & 0x3f) | 0x80;
    const value = [...bytes].map((byte) => byte.toString(16).padStart(2, "0")).join("");
    return `${value.slice(0, 8)}-${value.slice(8, 12)}-${value.slice(12, 16)}-${value.slice(16, 20)}-${value.slice(20)}`;
  };
}

document.querySelectorAll("[data-password-toggle]").forEach((button) => {
  button.addEventListener("click", () => {
    const input = document.getElementById(button.dataset.target);
    if (!input) return;

    const visible = input.type === "text";
    input.type = visible ? "password" : "text";
    button.setAttribute("aria-pressed", String(!visible));
    button.setAttribute("aria-label", visible ? "Tampilkan password" : "Sembunyikan password");
    const icon = button.querySelector("[data-password-icon]");
    icon?.classList.toggle("fa-eye", visible);
    icon?.classList.toggle("fa-eye-slash", !visible);
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

document.querySelectorAll("[data-sidebar-toggle]").forEach((button) => {
  button.addEventListener("click", () => {
    const parent = button.closest("[data-sidebar-parent]");
    const submenu = parent?.querySelector("[data-sidebar-submenu]");
    const chevron = button.querySelector("[data-sidebar-chevron]");
    const opening = submenu?.classList.contains("hidden") ?? false;
    submenu?.classList.toggle("hidden", !opening);
    chevron?.classList.toggle("rotate-180", opening);
    button.setAttribute("aria-expanded", String(opening));
  });
});

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
  if (recordCount) recordCount.textContent = `${visible} kasus`;
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

document.querySelectorAll("[data-history-back]").forEach((link) => {
  link.addEventListener("click", (event) => {
    if (!document.referrer || window.history.length <= 1) return;
    const previousURL = new URL(document.referrer);
    if (previousURL.origin !== window.location.origin) return;
    event.preventDefault();
    window.history.back();
  });
});

document.querySelectorAll("[data-open-dialog]").forEach((button) => {
  button.addEventListener("click", () => document.getElementById(button.dataset.openDialog)?.showModal());
});
document.querySelectorAll("[data-close-dialog]").forEach((button) => {
  button.addEventListener("click", () => button.closest("dialog")?.close());
});
document.querySelectorAll("dialog").forEach((dialog) => {
  dialog.addEventListener("click", (event) => {
    if (event.target === dialog) dialog.close();
  });
});

document.querySelectorAll("[data-edit-role]").forEach((button) => {
  button.addEventListener("click", () => {
    const dialog = document.getElementById("editRoleDialog");
    const form = document.getElementById("editRoleForm");
    if (!dialog || !form) return;
    form.action = `/settings/roles/${button.dataset.id}/update`;
    form.elements.name.value = button.dataset.name;
    form.elements.guard_name.value = button.dataset.guard;
    form.elements.description.value = button.dataset.description;
    dialog.showModal();
  });
});

document.querySelectorAll("[data-edit-permission]").forEach((button) => {
  button.addEventListener("click", () => {
    const dialog = document.getElementById("editPermissionDialog");
    const form = document.getElementById("editPermissionForm");
    if (!dialog || !form) return;
    form.action = `/settings/permissions/${button.dataset.id}/update`;
    form.elements.name.value = button.dataset.name;
    form.elements.guard_name.value = button.dataset.guard;
    form.elements.description.value = button.dataset.description;
    dialog.showModal();
  });
});

document.querySelectorAll("[data-edit-user]").forEach((button) => {
  button.addEventListener("click", () => {
    const dialog = document.getElementById("editUserDialog");
    const form = document.getElementById("editUserForm");
    if (!dialog || !form) return;
    form.action = `/settings/users/${button.dataset.id}/update`;
    form.elements.name.value = button.dataset.name ?? "";
    form.elements.employee_number.value = button.dataset.employeeNumber ?? "";
    form.elements.username.value = button.dataset.username ?? "";
    form.elements.email.value = button.dataset.email ?? "";
    form.elements.phone.value = button.dataset.phone ?? "";
    form.elements.password.value = "";
    form.elements.status.value = button.dataset.status ?? "active";
    form.elements.store_id.value = button.dataset.storeId ?? "";
    form.elements.role_id.value = button.dataset.roleId ?? "";
    dialog.showModal();
  });
});

document.querySelectorAll("[data-edit-store]").forEach((button) => {
  button.addEventListener("click", () => {
    const dialog = document.getElementById("editStoreDialog");
    const form = document.getElementById("editStoreForm");
    if (!dialog || !form) return;
    form.action = `/master/stores/${button.dataset.id}/update`;
    form.elements.code.value = button.dataset.code ?? "";
    form.elements.name.value = button.dataset.name ?? "";
    form.elements.address.value = button.dataset.address ?? "";
    form.elements.phone.value = button.dataset.phone ?? "";
    form.elements.timezone.value = button.dataset.timezone ?? "Asia/Jakarta";
    form.elements.status.value = button.dataset.status ?? "active";
    dialog.showModal();
  });
});

document.querySelectorAll("[data-edit-area]").forEach((button) => {
  button.addEventListener("click", () => {
    const dialog = document.getElementById("editAreaDialog");
    const form = document.getElementById("editAreaForm");
    if (!dialog || !form) return;
    form.action = `/master/areas/${button.dataset.id}/update`;
    form.elements.store_id.value = button.dataset.storeId ?? "";
    form.elements.code.value = button.dataset.code ?? "";
    form.elements.name.value = button.dataset.name ?? "";
    form.elements.sort_order.value = button.dataset.sortOrder ?? "0";
    form.elements.status.value = button.dataset.status ?? "active";
    form.elements.is_priority.checked = button.dataset.priority === "1";
    dialog.showModal();
  });
});

document.querySelectorAll("[data-edit-checklist]").forEach((button) => {
  button.addEventListener("click", () => {
    const dialog = document.getElementById("editChecklistDialog");
    const form = document.getElementById("editChecklistForm");
    if (!dialog || !form) return;
    form.action = `/master/checklists/${button.dataset.id}/update`;
    form.elements.code.value = button.dataset.code ?? "";
    form.elements.phase.value = button.dataset.phase ?? "start";
    form.elements.name.value = button.dataset.name ?? "";
    form.elements.sort_order.value = button.dataset.sortOrder ?? "0";
    form.elements.status.value = button.dataset.status ?? "active";
    form.elements.requires_camera_count.checked = button.dataset.requiresCameraCount === "1";
    dialog.showModal();
  });
});

document.querySelectorAll("[data-edit-incident-type]").forEach((button) => {
  button.addEventListener("click", () => {
    const dialog = document.getElementById("editIncidentTypeDialog");
    const form = document.getElementById("editIncidentTypeForm");
    if (!dialog || !form) return;
    form.action = `/master/incident-types/${button.dataset.id}/update`;
    form.elements.code.value = button.dataset.code ?? "";
    form.elements.name.value = button.dataset.name ?? "";
    form.elements.sort_order.value = button.dataset.sortOrder ?? "0";
    form.elements.status.value = button.dataset.status ?? "active";
    form.elements.requires_description.checked = button.dataset.requiresDescription === "1";
    dialog.showModal();
  });
});

function filterCameraAreas(form, selectedArea = "") {
  const store = form?.querySelector("[data-camera-store]");
  const area = form?.querySelector("[data-camera-area]");
  if (!store || !area) return;
  const storeID = store.value;
  let available = 0;
  [...area.options].forEach((option, index) => {
    if (index === 0) return;
    const visible = Boolean(storeID) && option.dataset.storeId === storeID;
    option.hidden = !visible;
    option.disabled = !visible;
    if (visible) available += 1;
  });
  area.disabled = !storeID || available === 0;
  area.options[0].textContent = !storeID ? "Pilih store terlebih dahulu" : available ? "Pilih area" : "Belum ada area pada store ini";
  area.value = selectedArea && [...area.options].some((option) => !option.disabled && option.value === selectedArea) ? selectedArea : "";
}

document.querySelectorAll("[data-camera-form]").forEach((form) => {
  form.querySelector("[data-camera-store]")?.addEventListener("change", () => filterCameraAreas(form));
  filterCameraAreas(form);
});

document.querySelectorAll("[data-edit-camera]").forEach((button) => {
  button.addEventListener("click", () => {
    const dialog = document.getElementById("editCameraDialog");
    const form = document.getElementById("editCameraForm");
    if (!dialog || !form) return;
    form.action = `/master/cameras/${button.dataset.id}/update`;
    form.elements.store_id.value = button.dataset.storeId ?? "";
    filterCameraAreas(form, button.dataset.areaId ?? "");
    form.elements.code.value = button.dataset.code ?? "";
    form.elements.name.value = button.dataset.name ?? "";
    form.elements.description.value = button.dataset.description ?? "";
    form.elements.status.value = button.dataset.status ?? "active";
    dialog.showModal();
  });
});

document.querySelectorAll("[data-edit-shift]").forEach((button) => {
  button.addEventListener("click", () => {
    const dialog = document.getElementById("editShiftDialog");
    const form = document.getElementById("editShiftForm");
    if (!dialog || !form) return;
    form.action = `/master/shifts/${button.dataset.id}/update`;
    form.elements.code.value = button.dataset.code ?? "";
    form.elements.name.value = button.dataset.name ?? "";
    form.elements.start_time.value = button.dataset.startTime ?? "";
    form.elements.end_time.value = button.dataset.endTime ?? "";
    form.elements.sort_order.value = button.dataset.sortOrder ?? "0";
    form.elements.status.value = button.dataset.status ?? "active";
    dialog.showModal();
  });
});

document.querySelectorAll("[data-edit-action]").forEach((button) => {
  button.addEventListener("click", () => {
    const dialog = document.getElementById("editActionDialog");
    const form = document.getElementById("editActionForm");
    if (!dialog || !form) return;
    form.action = `/master/actions/${button.dataset.id}/update`;
    form.elements.code.value = button.dataset.code ?? "";
    form.elements.name.value = button.dataset.name ?? "";
    form.elements.sort_order.value = button.dataset.sortOrder ?? "0";
    form.elements.status.value = button.dataset.status ?? "active";
    form.elements.requires_description.checked = button.dataset.requiresDescription === "1";
    dialog.showModal();
  });
});

document.querySelectorAll("[data-user-store-row]").forEach((row) => {
  const checkbox = row.querySelector("[data-user-store-checkbox]");
  const role = row.querySelector("[data-user-store-role]");
  const defaultRadio = row.querySelector("[data-user-store-default]");
  checkbox?.addEventListener("change", () => {
    const enabled = checkbox.checked;
    if (role) role.disabled = !enabled;
    if (defaultRadio) {
      defaultRadio.disabled = !enabled;
      if (!enabled) defaultRadio.checked = false;
    }
    if (enabled && !document.querySelector("[data-user-store-default]:checked")) defaultRadio.checked = true;
  });
});

document.querySelectorAll("form[data-confirm]").forEach((form) => {
  const deleteButton = form.querySelector("button");
  if (deleteButton && !deleteButton.title) deleteButton.title = deleteButton.disabled ? "Data dilindungi" : "Hapus";
  form.addEventListener("submit", async (event) => {
    event.preventDefault();

    if (typeof Swal === "undefined") {
      if (window.confirm(form.dataset.confirm)) form.submit();
      return;
    }

    const result = await Swal.fire({
      title: "Hapus data?",
      text: form.dataset.confirm,
      icon: "warning",
      showCancelButton: true,
      confirmButtonText: "Ya, hapus",
      cancelButtonText: "Batal",
      confirmButtonColor: "#ef4444",
      cancelButtonColor: "#64748b",
      reverseButtons: true,
      focusCancel: true,
    });

    if (result.isConfirmed) form.submit();
  });
});

document.querySelectorAll("[data-edit-role], [data-edit-permission], [data-edit-user], [data-edit-store], [data-edit-area], [data-edit-checklist], [data-edit-incident-type], [data-edit-camera], [data-edit-shift], [data-edit-action]").forEach((button) => {
  button.title = "Edit";
});

document.querySelectorAll("[data-access-filter]").forEach((filter) => {
  const table = filter.nextElementSibling;
  if (!table?.matches("[data-filter-table]")) return;

  const search = filter.querySelector("[data-filter-search]");
  const guard = filter.querySelector("[data-filter-guard]");
  const store = filter.querySelector("[data-filter-store]");
  const count = filter.querySelector("[data-filter-count]");
  const rows = [...table.querySelectorAll("[data-filter-row]")];
  const empty = table.querySelector("[data-filter-empty]");
  const pageSize = 25;
  let currentPage = 1;

  const pagination = document.createElement("nav");
  pagination.setAttribute("aria-label", "Navigasi halaman tabel");
  pagination.className = "mt-4 flex flex-col gap-3 rounded-2xl border border-slate-200 bg-white px-4 py-3 shadow-sm sm:flex-row sm:items-center sm:justify-between";
  pagination.innerHTML = '<p data-pagination-info class="text-center text-xs text-slate-500 sm:text-left"></p><div data-pagination-buttons class="flex flex-wrap items-center justify-center gap-1.5"></div>';
  table.insertAdjacentElement("afterend", pagination);
  const paginationInfo = pagination.querySelector("[data-pagination-info]");
  const paginationButtons = pagination.querySelector("[data-pagination-buttons]");

  const existingOptions = new Set([...guard?.options ?? []].map((option) => option.value));
  [...new Set(rows.map((row) => row.dataset.filterGuard).filter((value) => value && !existingOptions.has(value)))]
    .sort((left, right) => left.localeCompare(right))
    .forEach((value) => guard?.add(new Option(value, value)));

  function renderPagination(totalItems, totalPages) {
    const start = totalItems === 0 ? 0 : ((currentPage - 1) * pageSize) + 1;
    const end = Math.min(currentPage * pageSize, totalItems);
    if (paginationInfo) paginationInfo.textContent = `Menampilkan ${start}–${end} dari ${totalItems} data`;
    if (!paginationButtons) return;
    paginationButtons.replaceChildren();

    const makeButton = (label, page, options = {}) => {
      const button = document.createElement("button");
      button.type = "button";
      button.disabled = options.disabled ?? false;
      button.setAttribute("aria-label", options.ariaLabel ?? `Halaman ${page}`);
      if (options.current) button.setAttribute("aria-current", "page");
      button.className = options.current
        ? "grid min-w-9 place-items-center rounded-lg bg-orange-500 px-2.5 py-2 text-xs font-bold text-white shadow-sm"
        : "grid min-w-9 place-items-center rounded-lg border border-slate-200 bg-white px-2.5 py-2 text-xs font-bold text-slate-600 transition hover:border-orange-200 hover:bg-orange-50 hover:text-orange-700 disabled:cursor-not-allowed disabled:opacity-40";
      button.innerHTML = label;
      button.addEventListener("click", () => { currentPage = page; applyAccessFilter(false); });
      return button;
    };

    paginationButtons.append(makeButton('<i class="fa-solid fa-chevron-left" aria-hidden="true"></i>', Math.max(1, currentPage - 1), { disabled: currentPage === 1, ariaLabel: "Halaman sebelumnya" }));
    const pages = [];
    for (let page = 1; page <= totalPages; page += 1) {
      if (page === 1 || page === totalPages || Math.abs(page - currentPage) <= 1) pages.push(page);
    }
    let previousPage = 0;
    pages.forEach((page) => {
      if (page - previousPage > 1) {
        const separator = document.createElement("span");
        separator.className = "px-1 text-xs text-slate-400";
        separator.textContent = "…";
        paginationButtons.append(separator);
      }
      paginationButtons.append(makeButton(String(page), page, { current: page === currentPage }));
      previousPage = page;
    });
    paginationButtons.append(makeButton('<i class="fa-solid fa-chevron-right" aria-hidden="true"></i>', Math.min(totalPages, currentPage + 1), { disabled: currentPage === totalPages, ariaLabel: "Halaman berikutnya" }));
  }

  function applyAccessFilter(resetPage = false) {
    if (resetPage) currentPage = 1;
    const query = search?.value.trim().toLowerCase() ?? "";
    const selectedGuard = guard?.value ?? "";
    const filteredRows = rows.filter((row) => {
      const matchesSearch = (row.dataset.filterSearch ?? "").toLowerCase().includes(query);
      const matchesGuard = !selectedGuard || row.dataset.filterGuard === selectedGuard;
      const matchesStore = !store?.value || row.dataset.filterStore === store.value;
      return matchesSearch && matchesGuard && matchesStore;
    });
    const totalPages = Math.max(1, Math.ceil(filteredRows.length / pageSize));
    currentPage = Math.min(currentPage, totalPages);
    const pageStart = (currentPage - 1) * pageSize;
    const pageRows = new Set(filteredRows.slice(pageStart, pageStart + pageSize));
    rows.forEach((row) => row.classList.toggle("hidden", !pageRows.has(row)));
    filteredRows.forEach((row, index) => {
      const number = row.querySelector("[data-row-number]");
      if (number) number.textContent = String(index + 1);
    });

    if (count) count.textContent = `${filteredRows.length} data`;
    empty?.classList.toggle("hidden", filteredRows.length !== 0 || rows.length === 0);
    renderPagination(filteredRows.length, totalPages);
  }

  search?.addEventListener("input", () => applyAccessFilter(true));
  guard?.addEventListener("change", () => applyAccessFilter(true));
  store?.addEventListener("change", () => applyAccessFilter(true));
  applyAccessFilter();
});

document.querySelectorAll("[data-permission-select]").forEach((button) => {
  button.addEventListener("click", () => {
    const checked = button.dataset.permissionSelect === "all";
    document.querySelectorAll('input[name="permission_ids"]').forEach((input) => { input.checked = checked; });
    updatePermissionCounters();
  });
});

function updatePermissionCounters() {
  const inputs = [...document.querySelectorAll('input[name="permission_ids"]')];
  const total = inputs.filter((input) => input.checked).length;
  const totalElement = document.getElementById("assignedPermissionCount");
  if (totalElement) totalElement.textContent = String(total);

  document.querySelectorAll("[data-permission-group]").forEach((group) => {
    const assigned = [...group.querySelectorAll('input[name="permission_ids"]')].filter((input) => input.checked).length;
    const counter = group.querySelector("[data-group-assigned]");
    if (counter) counter.textContent = String(assigned);
  });
}

document.querySelectorAll('input[name="permission_ids"]').forEach((input) => {
  input.addEventListener("change", updatePermissionCounters);
});

document.querySelectorAll("[data-group-select]").forEach((button) => {
  button.addEventListener("click", () => {
    const group = button.closest("[data-permission-group]");
    const checked = button.dataset.groupSelect === "all";
    group?.querySelectorAll('input[name="permission_ids"]').forEach((input) => { input.checked = checked; });
    updatePermissionCounters();
  });
});

document.getElementById("permissionSearch")?.addEventListener("input", (event) => {
  const query = event.target.value.trim().toLowerCase();
  let visibleTotal = 0;
  document.querySelectorAll("[data-permission-group]").forEach((group) => {
    let groupVisible = 0;
    group.querySelectorAll("[data-permission-item]").forEach((item) => {
      const visible = (item.dataset.search ?? "").toLowerCase().includes(query);
      item.classList.toggle("hidden", !visible);
      if (visible) groupVisible += 1;
    });
    group.classList.toggle("hidden", groupVisible === 0);
    visibleTotal += groupVisible;
  });
  document.getElementById("permissionEmpty")?.classList.toggle("hidden", visibleTotal !== 0);
});

const profileMenuButton = document.querySelector("[data-profile-menu-button]");
const profileMenu = document.querySelector("[data-profile-menu]");
const notificationWrapper = document.querySelector("[data-notification-wrapper]");
const notificationButton = document.querySelector("[data-notification-button]");
const notificationMenu = document.querySelector("[data-notification-menu]");

notificationWrapper?.addEventListener("toggle", () => {
  if (!notificationWrapper.open) return;
  profileMenu?.classList.add("hidden");
  profileMenuButton?.setAttribute("aria-expanded", "false");
});

document.querySelector("[data-notification-read-all]")?.addEventListener("click", () => {
  document.querySelectorAll("[data-notification-dot]").forEach((dot) => dot.classList.add("hidden"));
  document.querySelector("[data-notification-badge]")?.classList.add("hidden");
  showDashboardToast("Semua notifikasi ditandai sudah dibaca.");
});

profileMenuButton?.addEventListener("click", (event) => {
  event.stopPropagation();
  const opening = profileMenu?.classList.contains("hidden");
  profileMenu?.classList.toggle("hidden", !opening);
  profileMenuButton.setAttribute("aria-expanded", String(opening));
  notificationWrapper?.removeAttribute("open");
});

document.addEventListener("click", (event) => {
  if (profileMenu && profileMenuButton && !profileMenu.contains(event.target) && !profileMenuButton.contains(event.target)) {
    profileMenu.classList.add("hidden");
    profileMenuButton.setAttribute("aria-expanded", "false");
  }
  if (notificationWrapper && !notificationWrapper.contains(event.target)) notificationWrapper.removeAttribute("open");
});

const incidentList = document.querySelector("[data-incident-list]");
const incidentTemplate = document.getElementById("incidentFormTemplate");

function reindexIncidents() {
  const cards = [...(incidentList?.querySelectorAll("[data-incident-card]") ?? [])];
  cards.forEach((card, index) => {
    const title = card.querySelector("[data-incident-title]");
    if (title) title.textContent = `Insiden #${index + 1}`;
    const indexField = card.querySelector("[data-incident-index]");
    if (indexField) indexField.value = String(index);
    card.querySelectorAll("[name]").forEach((field) => {
      field.name = field.name.replace(/incident_\d+_/, `incident_${index}_`);
    });
  });
}

document.querySelector("[data-add-incident]")?.addEventListener("click", () => {
  if (!incidentList || !(incidentTemplate instanceof HTMLTemplateElement)) return;
  const card = incidentTemplate.content.firstElementChild?.cloneNode(true);
  if (!(card instanceof HTMLElement)) return;
  incidentList.append(card);
  reindexIncidents();
  card.scrollIntoView({ behavior: "smooth", block: "center" });
});

incidentList?.addEventListener("click", (event) => {
  const removeButton = event.target.closest("[data-remove-incident]");
  if (!removeButton) return;
  removeButton.closest("[data-incident-card]")?.remove();
  reindexIncidents();
});

reindexIncidents();

document.querySelectorAll("[data-signature-pad]").forEach((canvas) => {
  const wrapper = canvas.closest("div.rounded-2xl");
  const hidden = wrapper?.querySelector("[data-signature-value]");
  const clearButton = wrapper?.querySelector("[data-clear-signature]");
  const ratio = Math.max(window.devicePixelRatio || 1, 1);
  const width = Math.max(canvas.clientWidth, 280);
  const height = 160;
  canvas.width = Math.round(width * ratio);
  canvas.height = Math.round(height * ratio);
  const context = canvas.getContext("2d");
  context.scale(ratio, ratio);
  context.lineCap = "round";
  context.lineJoin = "round";
  context.lineWidth = 2.2;
  context.strokeStyle = "#0f172a";
  let drawing = false;

  const point = (event) => {
    const rect = canvas.getBoundingClientRect();
    return { x: event.clientX - rect.left, y: event.clientY - rect.top };
  };
  canvas.addEventListener("pointerdown", (event) => {
    drawing = true;
    canvas.setPointerCapture(event.pointerId);
    const current = point(event);
    context.beginPath();
    context.moveTo(current.x, current.y);
  });
  canvas.addEventListener("pointermove", (event) => {
    if (!drawing) return;
    const current = point(event);
    context.lineTo(current.x, current.y);
    context.stroke();
    canvas.dataset.signed = "true";
  });
  const stop = () => { drawing = false; context.closePath(); };
  canvas.addEventListener("pointerup", stop);
  canvas.addEventListener("pointercancel", stop);
  clearButton?.addEventListener("click", () => {
    context.clearRect(0, 0, width, height);
    delete canvas.dataset.signed;
    if (hidden) hidden.value = "";
  });
});

document.querySelectorAll("form[data-signature-form]").forEach((form) => {
  form.addEventListener("submit", (event) => {
    const action = event.submitter?.value || "";
    const needsSignature = !["reject", "revision"].includes(action);
    const canvas = form.querySelector("[data-signature-pad]");
    const hidden = form.querySelector("[data-signature-value]");
    if (canvas && hidden) {
      hidden.value = canvas.dataset.signed === "true" ? canvas.toDataURL("image/png") : "";
    }
    if (needsSignature && canvas?.dataset.signed !== "true") {
      event.preventDefault();
      showDashboardToast("Tanda tangan wajib diisi sebelum melanjutkan.");
      canvas.scrollIntoView({ behavior: "smooth", block: "center" });
    }
  });
});

document.getElementById("form-laporan")?.addEventListener("submit", (event) => {
  const pads = [...document.querySelectorAll("[data-signature-pad]")];
  pads.forEach((canvas) => {
    const hidden = canvas.closest("div.rounded-2xl")?.querySelector("[data-signature-value]");
    if (hidden) hidden.value = canvas.dataset.signed === "true" ? canvas.toDataURL("image/png") : "";
  });
  if (event.submitter?.value === "submitted" && pads.some((canvas) => canvas.dataset.signed !== "true")) {
    event.preventDefault();
    showDashboardToast("Tanda tangan petugas serah dan petugas terima wajib diisi sebelum laporan dikirim.");
    pads.find((canvas) => canvas.dataset.signed !== "true")?.scrollIntoView({ behavior: "smooth", block: "center" });
  }
});

const blmkcStore = document.getElementById("blmkcStore");
const blmkcScopedSelects = [...document.querySelectorAll("[data-blmkc-store-option]")];
function filterBLMKCStoreOptions(resetInvalid = false) {
  if (!blmkcStore) return;
  const storeID = blmkcStore.value;
  blmkcScopedSelects.forEach((select) => {
    [...select.options].forEach((option) => {
      if (!option.dataset.storeId) return;
      option.hidden = option.dataset.storeId !== storeID;
      option.disabled = option.hidden;
    });
    const selected = select.selectedOptions[0];
    if (resetInvalid && selected?.disabled) select.value = "";
  });
}
blmkcStore?.addEventListener("change", () => filterBLMKCStoreOptions(true));
filterBLMKCStoreOptions();

const existingSubject = document.querySelector('select[name="existing_subject_id"]');
if (existingSubject) {
  const subjectName = document.querySelector('input[name="subject_name"]');
  const initialStatus = document.querySelector('select[name="initial_status"]');
  const existingSubjectField = existingSubject.closest("[data-existing-subject-field]");
  const subjectNameField = subjectName?.closest("[data-subject-name-field]");
  const syncExistingSubject = () => {
    const identityKnown = initialStatus?.value === "identified";
    existingSubjectField?.classList.toggle("hidden", !identityKnown);
    existingSubject.disabled = !identityKnown;
    if (!identityKnown) existingSubject.value = "";
    const usingExisting = existingSubject.value !== "";
    if (subjectName) {
      subjectName.disabled = usingExisting;
      subjectName.required = identityKnown && !usingExisting;
      subjectNameField?.classList.toggle("opacity-50", usingExisting);
    }
  };
  existingSubject.addEventListener("change", syncExistingSubject);
  initialStatus?.addEventListener("change", syncExistingSubject);
  syncExistingSubject();
}

const evidenceCategory = document.querySelector('form[action$="/evidence"] select[name="category"]');
if (evidenceCategory) {
  const vehicleFields = document.createElement("div");
  vehicleFields.className = "hidden rounded-2xl border border-blue-200 bg-blue-50 p-4 md:col-span-2";
  vehicleFields.dataset.vehicleEvidenceFields = "";
  vehicleFields.innerHTML = `
    <div class="mb-4"><strong class="text-sm text-blue-900">Data kendaraan</strong><p class="mt-1 text-xs text-blue-700">Data ini akan tampil pada bagian Kendaraan di detail kasus.</p></div>
    <div class="grid gap-4 sm:grid-cols-2">
      <label><span class="mb-2 block text-xs font-semibold text-slate-700">Jenis kendaraan <b class="text-red-500">*</b></span><input name="vehicle_type" maxlength="100" class="w-full rounded-xl border border-slate-200 bg-white px-3 py-3 text-sm" placeholder="Contoh: Sepeda motor"></label>
      <label><span class="mb-2 block text-xs font-semibold text-slate-700">Nomor polisi <b class="text-red-500">*</b></span><input name="vehicle_plate" maxlength="30" class="w-full rounded-xl border border-slate-200 bg-white px-3 py-3 text-sm uppercase" placeholder="Contoh: AB 1234 CD"></label>
      <label><span class="mb-2 block text-xs font-semibold text-slate-700">Merek</span><input name="vehicle_brand" maxlength="100" class="w-full rounded-xl border border-slate-200 bg-white px-3 py-3 text-sm" placeholder="Merek kendaraan"></label>
      <label><span class="mb-2 block text-xs font-semibold text-slate-700">Model</span><input name="vehicle_model" maxlength="100" class="w-full rounded-xl border border-slate-200 bg-white px-3 py-3 text-sm" placeholder="Model kendaraan"></label>
      <label><span class="mb-2 block text-xs font-semibold text-slate-700">Warna</span><input name="vehicle_color" maxlength="80" class="w-full rounded-xl border border-slate-200 bg-white px-3 py-3 text-sm" placeholder="Warna kendaraan"></label>
      <label><span class="mb-2 block text-xs font-semibold text-slate-700">Catatan kendaraan</span><input name="vehicle_notes" maxlength="2000" class="w-full rounded-xl border border-slate-200 bg-white px-3 py-3 text-sm" placeholder="Ciri khusus atau informasi tambahan"></label>
    </div>`;
  evidenceCategory.closest(".grid")?.append(vehicleFields);
  const syncVehicleFields = () => {
    const visible = evidenceCategory.value === "vehicle_photo";
    vehicleFields.classList.toggle("hidden", !visible);
    vehicleFields.querySelectorAll("input").forEach((input) => {
      input.disabled = !visible;
      input.required = visible && (input.name === "vehicle_type" || input.name === "vehicle_plate");
    });
  };
  evidenceCategory.addEventListener("change", syncVehicleFields);
  syncVehicleFields();
}

const statusDialog = document.querySelector("[data-status-dialog]");
document.querySelector("[data-status-dialog-open]")?.addEventListener("click", () => {
  statusDialog?.classList.remove("hidden");
  statusDialog?.classList.add("flex");
  statusDialog?.querySelector("input:not([type=hidden]), textarea")?.focus();
});
statusDialog?.querySelectorAll("[data-status-dialog-close]").forEach((button) => {
  button.addEventListener("click", () => {
    statusDialog.classList.add("hidden");
    statusDialog.classList.remove("flex");
  });
});
statusDialog?.addEventListener("click", (event) => {
  if (event.target === statusDialog) {
    statusDialog.classList.add("hidden");
    statusDialog.classList.remove("flex");
  }
});

// Show BLMKC evidence photos in a gallery modal without leaving the page.
const blmkcImagePreviewLinks = [...document.querySelectorAll(
  'a[target="_blank"][href*="/blmkc/"], a[target="_blank"][href^="/wanted/cases/"]'
)].filter((link) => link.querySelector("img"));
if (blmkcImagePreviewLinks.length) {
  const lightbox = document.createElement("div");
  lightbox.className = "fixed inset-0 z-[100] hidden items-center justify-center bg-slate-950/80 p-3 backdrop-blur-sm sm:p-6";
  lightbox.setAttribute("role", "dialog");
  lightbox.setAttribute("aria-modal", "true");
  lightbox.setAttribute("aria-labelledby", "blmkcImageTitle");
  lightbox.innerHTML = `
    <div data-blmkc-image-panel class="flex max-h-[calc(100dvh-1.5rem)] w-full max-w-6xl flex-col overflow-hidden rounded-2xl border-2 border-orange-400 bg-white shadow-2xl shadow-orange-950/30 sm:max-h-[calc(100dvh-3rem)]">
      <header class="relative z-20 flex shrink-0 items-center justify-between gap-4 border-b border-orange-400 bg-gradient-to-r from-orange-600 via-orange-500 to-amber-400 px-4 py-3 text-white sm:px-5">
        <div class="min-w-0">
          <h2 id="blmkcImageTitle" class="truncate text-sm font-bold text-white sm:text-base"><i class="fa-solid fa-image mr-2 text-orange-100"></i>Pratinjau Foto Bukti</h2>
          <p data-blmkc-image-count class="mt-0.5 text-[10px] font-semibold text-orange-100"></p>
        </div>
        <button type="button" data-blmkc-image-close class="grid size-10 shrink-0 place-items-center rounded-xl border border-white/30 bg-white/15 text-lg text-white transition hover:bg-white hover:text-orange-600" aria-label="Tutup pratinjau">
          <i class="fa-solid fa-xmark"></i>
        </button>
      </header>
      <div class="relative isolate flex min-h-0 flex-1 items-center justify-center overflow-hidden bg-slate-950 p-2 sm:p-4">
        <div data-blmkc-image-loading class="absolute inset-0 grid place-items-center text-white">
          <span class="text-center"><i class="fa-solid fa-circle-notch fa-spin text-2xl"></i><small class="mt-3 block text-xs text-slate-300">Memuat foto...</small></span>
        </div>
        <img data-blmkc-image-full draggable="false" class="hidden max-h-full max-w-full touch-none select-none rounded-lg object-contain transition-transform duration-200" alt="">
        <button type="button" data-blmkc-image-prev class="absolute left-3 z-10 grid size-10 place-items-center rounded-full bg-black/60 text-white shadow-lg transition hover:bg-orange-500 sm:left-5" aria-label="Foto sebelumnya"><i class="fa-solid fa-chevron-left"></i></button>
        <button type="button" data-blmkc-image-next class="absolute right-3 z-10 grid size-10 place-items-center rounded-full bg-black/60 text-white shadow-lg transition hover:bg-orange-500 sm:right-5" aria-label="Foto berikutnya"><i class="fa-solid fa-chevron-right"></i></button>
        <div class="absolute bottom-3 left-1/2 z-10 flex -translate-x-1/2 items-center overflow-hidden rounded-xl border border-white/15 bg-black/70 text-white shadow-xl backdrop-blur-sm sm:bottom-5">
          <button type="button" data-blmkc-zoom-out class="grid size-10 place-items-center transition hover:bg-white/15 disabled:cursor-not-allowed disabled:opacity-40" aria-label="Perkecil foto"><i class="fa-solid fa-minus"></i></button>
          <button type="button" data-blmkc-zoom-reset class="h-10 min-w-16 border-x border-white/15 px-3 text-xs font-bold transition hover:bg-white/15" aria-label="Kembalikan ukuran foto">100%</button>
          <button type="button" data-blmkc-zoom-in class="grid size-10 place-items-center transition hover:bg-white/15 disabled:cursor-not-allowed disabled:opacity-40" aria-label="Perbesar foto"><i class="fa-solid fa-plus"></i></button>
        </div>
      </div>
      <footer class="relative z-20 shrink-0 border-t border-orange-200 bg-orange-50 px-4 py-3 sm:px-5">
        <p data-blmkc-image-caption class="text-xs font-medium leading-5 text-orange-950"></p>
      </footer>
    </div>
  `;
  document.body.append(lightbox);

  const panel = lightbox.querySelector("[data-blmkc-image-panel]");
  const fullImage = lightbox.querySelector("[data-blmkc-image-full]");
  const loading = lightbox.querySelector("[data-blmkc-image-loading]");
  const count = lightbox.querySelector("[data-blmkc-image-count]");
  const caption = lightbox.querySelector("[data-blmkc-image-caption]");
  const closeButton = lightbox.querySelector("[data-blmkc-image-close]");
  const previousButton = lightbox.querySelector("[data-blmkc-image-prev]");
  const nextButton = lightbox.querySelector("[data-blmkc-image-next]");
  const zoomOutButton = lightbox.querySelector("[data-blmkc-zoom-out]");
  const zoomResetButton = lightbox.querySelector("[data-blmkc-zoom-reset]");
  const zoomInButton = lightbox.querySelector("[data-blmkc-zoom-in]");
  let trigger = null;
  let activeIndex = 0;
  let zoomLevel = 1;
  let panX = 0;
  let panY = 0;
  let dragging = false;
  let dragStartX = 0;
  let dragStartY = 0;

  const renderImageTransform = () => {
    fullImage.style.transform = `translate3d(${panX}px, ${panY}px, 0) scale(${zoomLevel})`;
    fullImage.style.cursor = zoomLevel > 1 ? (dragging ? "grabbing" : "grab") : "zoom-in";
  };

  const setZoom = (level) => {
    zoomLevel = Math.min(3, Math.max(0.5, level));
    if (zoomLevel <= 1) {
      panX = 0;
      panY = 0;
    }
    renderImageTransform();
    zoomResetButton.textContent = `${Math.round(zoomLevel * 100)}%`;
    zoomOutButton.disabled = zoomLevel <= 0.5;
    zoomInButton.disabled = zoomLevel >= 3;
  };

  const showImage = (index) => {
    activeIndex = (index + blmkcImagePreviewLinks.length) % blmkcImagePreviewLinks.length;
    const link = blmkcImagePreviewLinks[activeIndex];
    const thumbnail = link.querySelector("img");
    const description = link.closest("article")?.querySelector("p")?.textContent?.trim();
    const alt = thumbnail?.alt?.trim() || "Foto bukti BLMKC";
    loading.innerHTML = '<span class="text-center"><i class="fa-solid fa-circle-notch fa-spin text-2xl"></i><small class="mt-3 block text-xs text-slate-300">Memuat foto...</small></span>';
    loading.classList.remove("hidden");
    fullImage.classList.add("hidden");
    setZoom(1);
    fullImage.alt = alt;
    fullImage.src = link.href;
    caption.textContent = description || alt;
    count.textContent = `Foto ${activeIndex + 1} dari ${blmkcImagePreviewLinks.length}`;
    const hasMultipleImages = blmkcImagePreviewLinks.length > 1;
    previousButton.classList.toggle("hidden", !hasMultipleImages);
    nextButton.classList.toggle("hidden", !hasMultipleImages);
  };

  const closeLightbox = () => {
    lightbox.classList.add("hidden");
    lightbox.classList.remove("flex");
    document.body.classList.remove("overflow-hidden");
    fullImage.removeAttribute("src");
    trigger?.focus();
  };

  blmkcImagePreviewLinks.forEach((link) => {
    link.addEventListener("click", (event) => {
      event.preventDefault();
      trigger = link;
      showImage(blmkcImagePreviewLinks.indexOf(link));
      lightbox.classList.remove("hidden");
      lightbox.classList.add("flex");
      document.body.classList.add("overflow-hidden");
      closeButton.focus();
    });
  });

  fullImage.addEventListener("load", () => {
    loading.classList.add("hidden");
    fullImage.classList.remove("hidden");
  });
  fullImage.addEventListener("error", () => {
    fullImage.classList.add("hidden");
    loading.innerHTML = '<span class="text-center text-red-300"><i class="fa-solid fa-circle-exclamation text-3xl"></i><small class="mt-3 block text-xs">Foto tidak dapat dimuat.</small></span>';
  });
  previousButton.addEventListener("click", () => showImage(activeIndex - 1));
  nextButton.addEventListener("click", () => showImage(activeIndex + 1));
  zoomOutButton.addEventListener("click", () => setZoom(zoomLevel - 0.25));
  zoomResetButton.addEventListener("click", () => setZoom(1));
  zoomInButton.addEventListener("click", () => setZoom(zoomLevel + 0.25));
  fullImage.addEventListener("dblclick", () => setZoom(zoomLevel > 1 ? 1 : 2));
  fullImage.addEventListener("pointerdown", (event) => {
    if (zoomLevel <= 1) return;
    event.preventDefault();
    dragging = true;
    dragStartX = event.clientX - panX;
    dragStartY = event.clientY - panY;
    fullImage.setPointerCapture(event.pointerId);
    fullImage.classList.remove("transition-transform");
    renderImageTransform();
  });
  fullImage.addEventListener("pointermove", (event) => {
    if (!dragging) return;
    panX = event.clientX - dragStartX;
    panY = event.clientY - dragStartY;
    renderImageTransform();
  });
  const stopImageDrag = (event) => {
    if (!dragging) return;
    dragging = false;
    if (fullImage.hasPointerCapture(event.pointerId)) fullImage.releasePointerCapture(event.pointerId);
    fullImage.classList.add("transition-transform");
    renderImageTransform();
  };
  fullImage.addEventListener("pointerup", stopImageDrag);
  fullImage.addEventListener("pointercancel", stopImageDrag);
  closeButton.addEventListener("click", closeLightbox);
  lightbox.addEventListener("click", (event) => {
    if (!panel.contains(event.target)) closeLightbox();
  });
  document.addEventListener("keydown", (event) => {
    if (lightbox.classList.contains("hidden")) return;
    if (event.key === "Escape") closeLightbox();
    if (event.key === "ArrowLeft") showImage(activeIndex - 1);
    if (event.key === "ArrowRight") showImage(activeIndex + 1);
    if (event.key === "+" || event.key === "=") setZoom(zoomLevel + 0.25);
    if (event.key === "-" || event.key === "_") setZoom(zoomLevel - 0.25);
    if (event.key === "0") setZoom(1);
  });
}
