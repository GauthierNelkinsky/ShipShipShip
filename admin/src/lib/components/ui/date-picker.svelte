<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { Button } from "$lib/components/ui";
    import { Calendar, ChevronLeft, ChevronRight } from "lucide-svelte";

    const dispatch = createEventDispatcher();

    export let value = "";
    export let placeholder = "Pick a date";
    export let disabled = false;
    export let includeTime = false;

    let showCalendar = false;
    let currentDate = new Date();
    let selectedDate: Date | null = null;
    let hours = "";
    let minutes = "";
    let buttonElement: HTMLElement | null = null;
    let calendarTop = 0;
    let calendarLeft = 0;

    // Watch for value changes and update selectedDate
    $: {
        if (value) {
            // Parse date and optionally time
            if (includeTime && value.includes("T")) {
                const [datePart, timePart] = value.split("T");
                selectedDate = new Date(`${datePart}T12:00:00Z`);
                const [h, m] = timePart.split(":");
                hours = h || "";
                minutes = m || "";
            } else {
                // Create date at noon UTC to avoid timezone issues
                selectedDate = new Date(`${value}T12:00:00Z`);
                hours = "";
                minutes = "";
            }
        } else {
            selectedDate = null;
            hours = "";
            minutes = "";
        }
    }

    // Set current date to show the month of selected date or today
    $: if (selectedDate) {
        currentDate = new Date(
            selectedDate.getFullYear(),
            selectedDate.getMonth(),
            1,
        );
    }

    const months = [
        "January",
        "February",
        "March",
        "April",
        "May",
        "June",
        "July",
        "August",
        "September",
        "October",
        "November",
        "December",
    ];

    const daysOfWeek = ["Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"];

    function getDaysInMonth(date: Date) {
        const year = date.getFullYear();
        const month = date.getMonth();
        const firstDay = new Date(year, month, 1);
        const lastDay = new Date(year, month + 1, 0);
        const daysInMonth = lastDay.getDate();
        const startDate = firstDay.getDay();

        const days = [];

        // Add empty cells for days before the first day of the month
        for (let i = 0; i < startDate; i++) {
            days.push(null);
        }

        // Add days of the month
        for (let i = 1; i <= daysInMonth; i++) {
            days.push(new Date(year, month, i));
        }

        return days;
    }

    function selectDate(date: Date) {
        // Get year, month and day in LOCAL timezone
        const year = date.getFullYear();
        const month = String(date.getMonth() + 1).padStart(2, "0");
        const day = String(date.getDate()).padStart(2, "0");

        // Store date in YYYY-MM-DD format - this avoids any timezone issues
        let dateStr = `${year}-${month}-${day}`;

        // If includeTime is enabled and we have time values, append them
        if (includeTime && hours && minutes) {
            dateStr += `T${hours.padStart(2, "0")}:${minutes.padStart(2, "0")}`;
        }

        value = dateStr;

        // selectedDate will be updated via the reactive statement
        if (!includeTime) {
            showCalendar = false;
        }
        dispatch("change", value);
    }

    function updateTime() {
        if (!selectedDate) return;

        const year = selectedDate.getFullYear();
        const month = String(selectedDate.getMonth() + 1).padStart(2, "0");
        const day = String(selectedDate.getDate()).padStart(2, "0");

        let dateStr = `${year}-${month}-${day}`;

        // Auto-fill the other field when one is set
        if (hours !== "" && minutes === "") {
            minutes = "00";
        } else if (minutes !== "" && hours === "") {
            hours = "00";
        }

        // Validate and append time if both hours and minutes are provided
        const h = parseInt(hours) || 0;
        const m = parseInt(minutes) || 0;

        if (
            hours !== "" &&
            minutes !== "" &&
            h >= 0 &&
            h < 24 &&
            m >= 0 &&
            m < 60
        ) {
            dateStr += `T${String(h).padStart(2, "0")}:${String(m).padStart(2, "0")}`;
        }

        value = dateStr;
        dispatch("change", value);
    }

    function goToPreviousMonth() {
        currentDate = new Date(
            currentDate.getFullYear(),
            currentDate.getMonth() - 1,
            1,
        );
    }

    function goToNextMonth() {
        currentDate = new Date(
            currentDate.getFullYear(),
            currentDate.getMonth() + 1,
            1,
        );
    }

    function isToday(date: Date) {
        const today = new Date();
        return (
            date.getDate() === today.getDate() &&
            date.getMonth() === today.getMonth() &&
            date.getFullYear() === today.getFullYear()
        );
    }

    function isSameDate(date1: Date | null, date2: Date) {
        if (!date1) return false;
        return (
            date1.getDate() === date2.getDate() &&
            date1.getMonth() === date2.getMonth() &&
            date1.getFullYear() === date2.getFullYear()
        );
    }

    function handleToggle(e: Event) {
        e.stopPropagation();
        if (!showCalendar && buttonElement) {
            const rect = buttonElement.getBoundingClientRect();
            calendarTop = rect.top;
            calendarLeft = rect.left - 224 - 8; // 224px (w-56) + 8px margin
        }
        showCalendar = !showCalendar;
    }

    function handleOutsideClick(e: Event) {
        if (showCalendar && e.target instanceof Element) {
            const calendarElement = document.querySelector(
                ".date-picker-calendar",
            );
            if (calendarElement && !calendarElement.contains(e.target)) {
                showCalendar = false;
            }
        }
    }
</script>

<svelte:window on:click={handleOutsideClick} />

<div class="relative">
    <div bind:this={buttonElement}>
        <Button
            variant="outline"
            size="sm"
            on:click={handleToggle}
            class="h-6 text-xs border-dashed {selectedDate
                ? 'border-solid'
                : ''}"
            {disabled}
        >
            <Calendar class="h-3 w-3 mr-1" />
            {#if selectedDate}
                {new Date(
                    selectedDate.getFullYear(),
                    selectedDate.getMonth(),
                    selectedDate.getDate(),
                ).toLocaleDateString("en-US", {
                    month: "short",
                    day: "numeric",
                    year: "numeric",
                })}
                {#if includeTime && hours && minutes}
                    <span class="ml-1 text-muted-foreground">
                        {hours.padStart(2, "0")}:{minutes.padStart(2, "0")}
                    </span>
                {/if}
            {:else}
                {placeholder}
            {/if}
        </Button>
    </div>

    {#if showCalendar}
        <div
            class="date-picker-calendar fixed z-[9999] shadow-xl p-2 w-56 rounded-lg border border-border bg-card text-card-foreground"
            style="top: {calendarTop}px; left: {calendarLeft}px;"
            on:click|stopPropagation
            on:keydown={(e) => e.key === "Escape" && (showCalendar = false)}
            role="dialog"
            aria-label="Date picker"
            tabindex="-1"
        >
            <!-- Calendar Header -->
            <div class="flex items-center justify-between mb-2">
                <Button
                    variant="ghost"
                    size="sm"
                    on:click={goToPreviousMonth}
                    class="h-7 w-7 p-0"
                >
                    <ChevronLeft class="h-3.5 w-3.5" />
                </Button>

                <div class="text-xs font-medium">
                    {months[currentDate.getMonth()]}
                    {currentDate.getFullYear()}
                </div>

                <Button
                    variant="ghost"
                    size="sm"
                    on:click={goToNextMonth}
                    class="h-7 w-7 p-0"
                >
                    <ChevronRight class="h-3.5 w-3.5" />
                </Button>
            </div>

            <!-- Days of Week Header -->
            <div class="grid grid-cols-7 gap-0.5 mb-1">
                {#each daysOfWeek as day}
                    <div
                        class="text-center text-[10px] font-medium text-muted-foreground py-0.5"
                    >
                        {day}
                    </div>
                {/each}
            </div>

            <!-- Calendar Days -->
            <div class="grid grid-cols-7 gap-0.5">
                {#each getDaysInMonth(currentDate) as day}
                    {#if day}
                        <button
                            type="button"
                            class="h-7 w-7 text-center text-xs rounded hover:bg-accent hover:text-accent-foreground transition-colors {isSameDate(
                                selectedDate,
                                day,
                            )
                                ? 'bg-primary text-primary-foreground'
                                : isToday(day)
                                  ? 'bg-accent text-accent-foreground'
                                  : ''}"
                            on:click={() => selectDate(day)}
                        >
                            {day.getDate()}
                        </button>
                    {:else}
                        <div class="h-7 w-7"></div>
                    {/if}
                {/each}
            </div>

            <!-- Time Input (optional) -->
            {#if includeTime && selectedDate}
                <div class="mt-2 pt-2 border-t border-border">
                    <div
                        class="text-[10px] font-medium text-muted-foreground mb-1.5"
                    >
                        Time (optional)
                    </div>
                    <div class="flex items-center justify-center gap-1.5">
                        <input
                            type="number"
                            min="0"
                            max="23"
                            placeholder="HH"
                            bind:value={hours}
                            on:input={updateTime}
                            class="w-12 h-7 px-1.5 text-xs text-center rounded border border-input bg-background focus:outline-none focus:ring-1 focus:ring-ring"
                        />
                        <span class="text-xs text-muted-foreground">:</span>
                        <input
                            type="number"
                            min="0"
                            max="59"
                            placeholder="MM"
                            bind:value={minutes}
                            on:input={updateTime}
                            class="w-12 h-7 px-1.5 text-xs text-center rounded border border-input bg-background focus:outline-none focus:ring-1 focus:ring-ring"
                        />
                    </div>
                </div>
            {/if}

            <!-- Clear Button -->
            {#if selectedDate}
                <div class="mt-2 pt-2 border-t border-border">
                    <Button
                        variant="ghost"
                        size="sm"
                        on:click={() => {
                            selectedDate = null;
                            value = "";
                            hours = "";
                            minutes = "";
                            showCalendar = false;
                            dispatch("change", "");
                        }}
                        class="w-full h-7 text-xs"
                    >
                        Clear date
                    </Button>
                </div>
            {/if}
        </div>
    {/if}
</div>
