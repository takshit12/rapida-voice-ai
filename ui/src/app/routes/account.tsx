import { MissionBox } from '@/app/components/container/mission-box';
import { ProtectedBox } from '@/app/components/container/protected-box';
import { AccountSettingPage } from '@/app/pages/user';
import React from 'react';
import { Routes, Route } from 'react-router-dom';

export function AccountRoute() {
  return (
    <Routes>
      <Route
        key="setting"
        path=""
        element={
          <ProtectedBox>
            <MissionBox>
              <AccountSettingPage />
            </MissionBox>
          </ProtectedBox>
        }
      />
    </Routes>
  );
}
