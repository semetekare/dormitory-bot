package mock

import "github.com/google/uuid"

// UUIDs — экспортируемые идентификаторы мок-данных для использования в main.go.

// MockDormID8 возвращает UUID общежития №8 в мок-режиме.
func MockDormID8() uuid.UUID { return mockDormID8 }

// MockDormID5 возвращает UUID общежития №5 в мок-режиме.
func MockDormID5() uuid.UUID { return mockDormID5 }
