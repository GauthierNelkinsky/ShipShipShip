<script lang="ts">
    import { createEventDispatcher } from "svelte";
    import { Button } from "$lib/components/ui";
    import { Calendar, ChevronLeft, ChevronRight } from "lucide-svelte";

    const dispatch = createEventDispatcher();

    export let value = "";
    export let placeholder = "Pick a date";
    export let disabled = false;

    let showCalendar = false;
    let currentDate = new Date();
    let selectedDate = null;

    // Watch for value changes and update selectedDate
    $: {
        if (value) {
            // Create date at noon UTC to avoid timezone issues
            selectedDate = new Date(`${value}T12:00:00Z`);
        } else {
            selectedDate = null;
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
        const dateStr = `${year}-${month}-${day}`;
        value = dateStr;

        // selectedDate will be updated via the reactive statement
        showCalendar = false;
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
    <Button
        variant="outline"
        size="sm"
        on:click={handleToggle}
        class="h-6 text-xs border-dashed {selectedDate ? 'border-solid' : ''}"
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
        {:else}
            {placeholder}
        {/if}
    </Button>

    {#if showCalendar}
        <div
            class="date-picker-calendar absolute top-8 left-0 z-50 shadow-lg p-3 w-64 rounded-lg border border-border bg-card text-card-foreground"
            on:click|stopPropagation
            on:keydown={(e) => e.key === "Escape" && (showCalendar = false)}
            role="dialog"
            aria-label="Date picker"
            tabindex="-1"
        >
            <!-- Calendar Header -->
            <div class="flex items-center justify-between mb-4">
                <Button
                    variant="ghost"
                    size="sm"
                    on:click={goToPreviousMonth}
                    class="h-8 w-8 p-0"
                >
                    <ChevronLeft class="h-4 w-4" />
                </Button>

                <div class="text-sm font-medium">
                    {months[currentDate.getMonth()]}
                    {currentDate.getFullYear()}
                </div>

                <Button
                    variant="ghost"
                    size="sm"
                    on:click={goToNextMonth}
                    class="h-8 w-8 p-0"
                >
                    <ChevronRight class="h-4 w-4" />
                </Button>
            </div>

            <!-- Days of Week Header -->
            <div class="grid grid-cols-7 gap-1 mb-2">
                {#each daysOfWeek as day}
                    <div
                        class="text-center text-xs font-medium text-muted-foreground p-1"
                    >
                        {day}
                    </div>
                {/each}
            </div>

            <!-- Calendar Days -->
            <div class="grid grid-cols-7 gap-1">
                {#each getDaysInMonth(currentDate) as day}
                    {#if day}
                        <button
                            type="button"
                            class="h-8 w-8 text-center text-xs rounded-md hover:bg-accent hover:text-accent-foreground transition-colors {isSameDate(
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
                        <div class="h-8 w-8"></div>
                    {/if}
                {/each}
            </div>

            <!-- Clear Button -->
            {#if selectedDate}
                <div class="mt-3 pt-3 border-t border-border">
                    <Button
                        variant="ghost"
                        size="sm"
                        on:click={() => {
                            selectedDate = null;
                            value = "";
                            showCalendar = false;
                            dispatch("change", "");
                        }}
                        class="w-full text-xs"
                    >
                        Clear date
                    </Button>
                </div>
            {/if}
        </div>
    {/if}
</div>
