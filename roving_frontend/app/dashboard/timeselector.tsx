'use client';

import { DateRangePicker, DateRangePickerValue } from '@tremor/react';

interface TimeSelectorProps {
  value: DateRangePickerValue;
  setValue: (newValue: DateRangePickerValue) => void;
}

export default function TimeSelector({ value, setValue }: TimeSelectorProps) {
  return (
    <div className="max-w-sm mx-auto space-y-6">
      <DateRangePicker
        value={value}
        onValueChange={setValue}
        className="mt-4 max-w-sm mx-auto"
        maxDate={new Date()}
      />
    </div>
  );
}
