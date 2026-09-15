import React, { useState } from 'react';
import '../styles/Calendar.css';

const Calendar = () => {
    const [currentDate, setCurrentDate] = useState(new Date());
    const [selectedDate, setSelectedDate] = useState(null);

    const currentMonth = currentDate.getMonth();
    const currentYear = currentDate.getFullYear();
    const today = new Date();

    const getDaysInMonth = (month, year) => {
        return new Date(year, month + 1, 0).getDate();
    };

    const getFirstDayOfMonth = (month, year) => {
        return new Date(year, month, 1).getDay();
    };

    const prevMonth = () => {
        setCurrentDate(new Date(currentYear, currentMonth - 1, 1));
    };

    const nextMonth = () => {
        setCurrentDate(new Date(currentYear, currentMonth + 1, 1));
    };

    const goToToday = () => {
        setCurrentDate(new Date());
        setSelectedDate(null);
    };

    const generateCalendar = () => {
        const daysInMonth = getDaysInMonth(currentMonth, currentYear);
        const firstDay = getFirstDayOfMonth(currentMonth, currentYear);

        const daysInPrevMonth = getDaysInMonth(currentMonth - 1, currentYear);
        const prevMonthDays = [];
        for (let i = firstDay - 1; i >= 0; i--) {
            prevMonthDays.push({
                day: daysInPrevMonth - i,
                isCurrentMonth: false,
                month: currentMonth - 1,
                year: currentYear
            });
        }

        const currentMonthDays = [];
        for (let i = 1; i <= daysInMonth; i++) {
            currentMonthDays.push({
                day: i,
                isCurrentMonth: true,
                month: currentMonth,
                year: currentYear
            });
        }

        const totalCells = prevMonthDays.length + currentMonthDays.length;
        const nextMonthDays = [];
        const remainingCells = 42 - totalCells;
        for (let i = 1; i <= remainingCells; i++) {
            nextMonthDays.push({
                day: i,
                isCurrentMonth: false,
                month: currentMonth + 1,
                year: currentYear
            });
        }

        return [...prevMonthDays, ...currentMonthDays, ...nextMonthDays];
    };

    const calendarDays = generateCalendar();

    const monthNames = [
        'Январь', 'Февраль', 'Март', 'Апрель', 'Май', 'Июнь',
        'Июль', 'Август', 'Сентябрь', 'Октябрь', 'Ноябрь', 'Декабрь'
    ];

    const dayNames = ['Вс', 'Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб'];

    const isToday = (day, month, year) => {
        return today.getDate() === day &&
            today.getMonth() === month &&
            today.getFullYear() === year;
    };

    const isSelected = (day, month, year) => {
        if (!selectedDate) return false;
        return selectedDate.getDate() === day &&
            selectedDate.getMonth() === month &&
            selectedDate.getFullYear() === year;
    };

    const handleDateClick = (day, month, year, isCurrentMonth) => {
        const newDate = new Date(year, month, day);
        setSelectedDate(newDate);

        if (!isCurrentMonth) {
            setCurrentDate(new Date(year, month, 1));
        }
    };

    return (
        <>
            <div className="calendar-container">
                <div className="calendar-header">
                    <button className="nav-button" onClick={prevMonth}>
                        ‹
                    </button>

                    <h2 className="calendar-title">
                        {monthNames[currentMonth]} {currentYear}
                    </h2>

                    <button className="nav-button" onClick={nextMonth}>
                        ›
                    </button>
                </div>

                <div className="weekdays">
                    {dayNames.map(day => (
                        <div key={day} className="weekday">
                            {day}
                        </div>
                    ))}
                </div>

                <div className="calendar-grid">
                    {calendarDays.map((item, index) => {
                        const isTodayDate = isToday(item.day, item.month, item.year);
                        const isSelectedDate = isSelected(item.day, item.month, item.year);

                        let dayClass = 'calendar-day';
                        if (!item.isCurrentMonth) dayClass += ' other-month';
                        if (isTodayDate) dayClass += ' today';
                        if (isSelectedDate) dayClass += ' selected';

                        return (
                            <div
                                key={index}
                                className={dayClass}
                                onClick={() => handleDateClick(
                                    item.day,
                                    item.month,
                                    item.year,
                                    item.isCurrentMonth
                                )}
                            >
                                {item.day}
                            </div>
                        );
                    })}
                </div>

                <div className="calendar-footer">
                    <button className="today-button" onClick={goToToday}>
                        Сегодня
                    </button>
                </div>
            </div>
            {selectedDate && (
                <div className="selected-date">
                    <b>
                        {selectedDate.toLocaleDateString('ru-RU', {
                            year: 'numeric',
                            month: 'long',
                            day: 'numeric'
                        })}
                    </b>
                    <div className='selected-date-content'>
                        <span className='selected-content label'>Комната:</span>
                        <span className='selected-content value'>101</span>

                        <span className='selected-content label'>Вид дежурства:</span>
                        <span className='selected-content value'>Суточное</span>

                        <span className='selected-content label'>Место проведения:</span>
                        <span className='selected-content value'>Главный корпус</span>
                    </div>
                </div>
            )}
        </>
    );
};

export default Calendar;