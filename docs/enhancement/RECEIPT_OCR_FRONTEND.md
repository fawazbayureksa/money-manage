# Receipt OCR Frontend Implementation Plan

## Overview

Mobile-first interface for capturing receipt/bill images and reviewing extracted transaction data. The UX focuses on quick capture, easy error correction, and seamless confirmation.

---

## User Flow

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Capture    │────▶│  Processing  │────▶│   Review &   │────▶│  Transaction │
│   Receipt    │     │  (Loading)   │     │    Edit      │     │   Created!   │
└──────────────┘     └──────────────┘     └──────────────┘     └──────────────┘
       │                                         │
       │                                         │
       ▼                                         ▼
 • Camera capture                          • Edit extracted data
 • Gallery upload                          • Change category
 • Image preview                           • Select wallet/asset
                                          • Add notes
```

---

## Screens

### 1. Receipt Capture Screen

```
┌─────────────────────────────────────────┐
│ ←  Scan Receipt                         │
├─────────────────────────────────────────┤
│                                         │
│  ┌─────────────────────────────────┐   │
│  │                                 │   │
│  │                                 │   │
│  │                                 │   │
│  │                                 │   │
│  │      [ Camera Viewfinder ]      │   │
│  │                                 │   │
│  │                                 │   │
│  │   ┌─────────────────────────┐   │   │
│  │   │ Align receipt within    │   │   │
│  │   │ the frame               │   │   │
│  │   └─────────────────────────┘   │   │
│  │                                 │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ─────────────────────────────────────  │
│                                         │
│  💡 Tips for better results:            │
│  • Good lighting                        │
│  • Flat surface                         │
│  • Entire receipt visible               │
│                                         │
│  ┌────────┐  ┌────────────┐  ┌───────┐ │
│  │ 🖼️     │  │     📸     │  │  ⚡   │ │
│  │Gallery │  │  Capture   │  │ Flash │ │
│  └────────┘  └────────────┘  └───────┘ │
│                                         │
└─────────────────────────────────────────┘
```

**Features:**
- Camera viewfinder with guides
- Flash toggle
- Gallery picker
- Tips overlay

---

### 2. Image Preview & Upload

```
┌─────────────────────────────────────────┐
│ ←  Preview                      Retake  │
├─────────────────────────────────────────┤
│                                         │
│  ┌─────────────────────────────────┐   │
│  │                                 │   │
│  │                                 │   │
│  │                                 │   │
│  │                                 │   │
│  │     [ Captured Image ]          │   │
│  │                                 │   │
│  │     🔄 Tap to rotate            │   │
│  │     🔍 Pinch to zoom            │   │
│  │                                 │   │
│  │                                 │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ─────────────────────────────────────  │
│                                         │
│  Pre-select Wallet (Optional)           │
│  ┌─────────────────────────────────┐   │
│  │ 💳 BCA Checking            ▼   │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │                                 │   │
│  │         📤 Scan Receipt         │   │
│  │                                 │   │
│  └─────────────────────────────────┘   │
│                                         │
└─────────────────────────────────────────┘
```

**Features:**
- Image rotation
- Zoom/pan
- Retake button
- Optional wallet pre-selection
- Upload/scan button

---

### 3. Processing Screen

```
┌─────────────────────────────────────────┐
│                                         │
├─────────────────────────────────────────┤
│                                         │
│                                         │
│                                         │
│                                         │
│              ┌─────────┐                │
│              │         │                │
│              │  📄     │                │
│              │  ████   │                │
│              │  ████   │                │
│              │  ████   │                │
│              └─────────┘                │
│                                         │
│              ⟳ Scanning...              │
│                                         │
│         Reading receipt details         │
│                                         │
│         ████████████░░░░  65%           │
│                                         │
│                                         │
│  ─────────────────────────────────────  │
│                                         │
│  This usually takes 5-10 seconds        │
│                                         │
│                                         │
│              [ Cancel ]                 │
│                                         │
└─────────────────────────────────────────┘
```

**Animation Ideas:**
- Scanning line moving across document
- Pulsing document icon
- Progress percentage
- Fun facts about spending (while waiting)

---

### 4. Review & Edit Screen

```
┌─────────────────────────────────────────┐
│ ←  Review Transaction              ✓    │
├─────────────────────────────────────────┤
│                                         │
│  ┌─────────────────────────────────┐   │
│  │  🧾 Thumbnail    Tap to view >  │   │
│  └─────────────────────────────────┘   │
│                                         │
│  Confidence: ████████░░ 87%             │
│                                         │
│  ─────────────────────────────────────  │
│                                         │
│  EXTRACTED DETAILS                      │
│                                         │
│  Merchant                               │
│  ┌─────────────────────────────────┐   │
│  │ Indomaret                    ✏️ │   │
│  └─────────────────────────────────┘   │
│                                         │
│  Amount *                               │
│  ┌─────────────────────────────────┐   │
│  │ Rp 125,000                   ✏️ │   │
│  └─────────────────────────────────┘   │
│                                         │
│  Date *                                 │
│  ┌─────────────────────────────────┐   │
│  │ Feb 11, 2026                 📅 │   │
│  └─────────────────────────────────┘   │
│                                         │
│  Category *                             │
│  ┌─────────────────────────────────┐   │
│  │ 🛒 Groceries     Suggested!  ▼ │   │
│  └─────────────────────────────────┘   │
│                                         │
│  Wallet/Asset *                         │
│  ┌─────────────────────────────────┐   │
│  │ 💳 BCA Checking              ▼ │   │
│  └─────────────────────────────────┘   │
│                                         │
│  Notes (Optional)                       │
│  ┌─────────────────────────────────┐   │
│  │ Weekly grocery shopping         │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ─────────────────────────────────────  │
│                                         │
│  ITEMS DETECTED (5)             Expand ▼│
│                                         │
│  ┌─────────────────────────────────┐   │
│  │         💾 Save Transaction     │   │
│  └─────────────────────────────────┘   │
│                                         │
└─────────────────────────────────────────┘
```

**Features:**
- Confidence indicator
- Editable fields with clear indicators
- Suggested category badge
- Items expansion
- Required field markers

---

### 5. Items Detail View (Expanded)

```
┌─────────────────────────────────────────┐
│ ←  Receipt Items                        │
├─────────────────────────────────────────┤
│                                         │
│  5 items detected                       │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ Indomie Goreng                  │   │
│  │ 5 × Rp 3,500         Rp 17,500  │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ Aqua 600ml                      │   │
│  │ 2 × Rp 4,000          Rp 8,000  │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ Teh Botol Sosro                 │   │
│  │ 3 × Rp 5,000         Rp 15,000  │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ Richeese Nabati                 │   │
│  │ 4 × Rp 3,000         Rp 12,000  │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ Silver Queen                    │   │
│  │ 2 × Rp 12,000        Rp 24,000  │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ─────────────────────────────────────  │
│  Subtotal              Rp 76,500        │
│  Tax (PPN 11%)          Rp 8,415        │
│  ─────────────────────────────────────  │
│  Total                 Rp 84,915        │
│  ─────────────────────────────────────  │
│                                         │
│  ℹ️ Items are for reference only.       │
│     The total amount will be saved.     │
│                                         │
└─────────────────────────────────────────┘
```

---

### 6. Success Screen

```
┌─────────────────────────────────────────┐
│                                         │
├─────────────────────────────────────────┤
│                                         │
│                                         │
│                                         │
│                  ✅                      │
│                                         │
│         Transaction Created!            │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │                                 │   │
│  │  🛒 Groceries                   │   │
│  │                                 │   │
│  │  Indomaret                      │   │
│  │  Feb 11, 2026                   │   │
│  │                                 │   │
│  │  -Rp 125,000                    │   │
│  │  from BCA Checking              │   │
│  │                                 │   │
│  └─────────────────────────────────┘   │
│                                         │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │       📸 Scan Another           │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │       📋 View Transaction       │   │
│  └─────────────────────────────────┘   │
│                                         │
│              [ Done ]                   │
│                                         │
└─────────────────────────────────────────┘
```

---

### 7. Error/Low Confidence Screen

```
┌─────────────────────────────────────────┐
│ ←  Scan Result                          │
├─────────────────────────────────────────┤
│                                         │
│                  ⚠️                      │
│                                         │
│       Having trouble reading            │
│       this receipt                      │
│                                         │
│  Confidence: ██░░░░░░░░ 23%             │
│                                         │
│  ─────────────────────────────────────  │
│                                         │
│  Possible issues:                       │
│  • Image is blurry                      │
│  • Receipt is crumpled or faded         │
│  • Poor lighting conditions             │
│                                         │
│  ─────────────────────────────────────  │
│                                         │
│  What you can do:                       │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │     📸 Retake Photo             │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │     ✏️ Enter Manually           │   │
│  │     (with extracted text help)  │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ─────────────────────────────────────  │
│                                         │
│  [ View Raw Text ]                      │
│                                         │
│  INDOMARET                              │
│  JL SUDIRMAN NO 123                     │
│  11/02/2026 14:35                       │
│  ===============                        │
│  INDOMIE GRG    5x3.500      17.500     │
│  ...                         Show more  │
│                                         │
└─────────────────────────────────────────┘
```

---

### 8. Scan History Screen

```
┌─────────────────────────────────────────┐
│ ←  Receipt Scans                    ➕  │
├─────────────────────────────────────────┤
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ ▼ Pending Review (2)            │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ 🧾  Alfamart                    │   │
│  │     Rp 45,000 • Today 10:30     │   │
│  │     ⚡ 92% confidence    Review >│   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ 🧾  Unknown Merchant            │   │
│  │     Rp ? • Yesterday            │   │
│  │     ⚠️ 35% confidence   Review >│   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ ▼ Completed (15)                │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ ✓  Indomaret                    │   │
│  │     Rp 125,000 • Feb 11         │   │
│  │     Linked to transaction       │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ ✓  Starbucks                    │   │
│  │     Rp 85,000 • Feb 10          │   │
│  │     Linked to transaction       │   │
│  └─────────────────────────────────┘   │
│                                         │
└─────────────────────────────────────────┘
```

---

## Component Architecture

### React Native / Flutter Structure

```
screens/
├── receipt/
│   ├── ReceiptCaptureScreen.tsx
│   ├── ImagePreviewScreen.tsx
│   ├── ProcessingScreen.tsx
│   ├── ReviewEditScreen.tsx
│   ├── ItemsDetailScreen.tsx
│   ├── SuccessScreen.tsx
│   └── ScanHistoryScreen.tsx
│
components/
├── receipt/
│   ├── CameraViewfinder.tsx
│   ├── ConfidenceIndicator.tsx
│   ├── EditableField.tsx
│   ├── ReceiptThumbnail.tsx
│   ├── ItemsList.tsx
│   ├── ScanStatusBadge.tsx
│   └── ProcessingAnimation.tsx
│
hooks/
├── useCamera.ts
├── useReceiptScan.ts
└── useImagePicker.ts
│
services/
├── receiptApi.ts
└── imageService.ts
```

---

## API Integration

### Service Functions

```typescript
// services/receiptApi.ts

import api from './api';

export interface UploadReceiptRequest {
  image: File | Blob;
  assetId?: number;
}

export interface ExtractedData {
  merchant_name: string;
  date: string;
  time?: string;
  total_amount: number;
  subtotal?: number;
  tax?: number;
  items?: ReceiptItem[];
  payment_method?: string;
  suggested_category?: {
    id: number;
    name: string;
  };
}

export interface ReceiptItem {
  name: string;
  quantity: number;
  unit_price: number;
  total: number;
}

export interface ScanResult {
  scan_id: number;
  status: 'uploaded' | 'processing' | 'completed' | 'failed' | 'reviewed';
  image_url: string;
  extracted_data?: ExtractedData;
  confidence_score?: number;
  raw_text?: string;
  error_message?: string;
}

export interface ConfirmTransactionRequest {
  amount: number;
  date: string;
  description: string;
  category_id: number;
  asset_id: number;
  transaction_type: number;
  notes?: string;
}

// Upload receipt image
export const uploadReceipt = async (data: UploadReceiptRequest): Promise<{ scan_id: number; status: string }> => {
  const formData = new FormData();
  formData.append('image', data.image);
  if (data.assetId) {
    formData.append('asset_id', data.assetId.toString());
  }

  const response = await api.post('/receipts/scan', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
  return response.data.data;
};

// Get scan status/result
export const getScanResult = async (scanId: number): Promise<ScanResult> => {
  const response = await api.get(`/receipts/scan/${scanId}`);
  return response.data.data;
};

// Poll for result (with timeout)
export const pollScanResult = async (
  scanId: number,
  maxAttempts: number = 30,
  intervalMs: number = 2000
): Promise<ScanResult> => {
  for (let attempt = 0; attempt < maxAttempts; attempt++) {
    const result = await getScanResult(scanId);
    
    if (result.status === 'completed' || result.status === 'failed') {
      return result;
    }
    
    await new Promise(resolve => setTimeout(resolve, intervalMs));
  }
  
  throw new Error('Scan timeout - please try again');
};

// Confirm and create transaction
export const confirmReceipt = async (
  scanId: number,
  data: ConfirmTransactionRequest
): Promise<{ transaction_id: number }> => {
  const response = await api.post(`/receipts/scan/${scanId}/confirm`, data);
  return response.data.data;
};

// List user's scans
export const listScans = async (
  status?: string,
  page: number = 1,
  limit: number = 20
): Promise<{ data: ScanResult[]; total: number }> => {
  const params = new URLSearchParams({ page: page.toString(), limit: limit.toString() });
  if (status) params.append('status', status);
  
  const response = await api.get(`/receipts?${params}`);
  return response.data;
};

// Delete scan
export const deleteScan = async (scanId: number): Promise<void> => {
  await api.delete(`/receipts/scan/${scanId}`);
};
```

---

## State Management

### Receipt Scan Store (Zustand/Redux)

```typescript
// store/receiptStore.ts

import { create } from 'zustand';
import * as receiptApi from '../services/receiptApi';

interface ReceiptState {
  // Current scan
  currentScan: receiptApi.ScanResult | null;
  isUploading: boolean;
  isProcessing: boolean;
  error: string | null;
  
  // Form data for review
  editedData: Partial<receiptApi.ConfirmTransactionRequest>;
  
  // Actions
  uploadReceipt: (image: File | Blob, assetId?: number) => Promise<void>;
  pollForResult: (scanId: number) => Promise<void>;
  updateEditedData: (data: Partial<receiptApi.ConfirmTransactionRequest>) => void;
  confirmTransaction: (scanId: number) => Promise<number>;
  reset: () => void;
}

export const useReceiptStore = create<ReceiptState>((set, get) => ({
  currentScan: null,
  isUploading: false,
  isProcessing: false,
  error: null,
  editedData: {},
  
  uploadReceipt: async (image, assetId) => {
    set({ isUploading: true, error: null });
    try {
      const result = await receiptApi.uploadReceipt({ image, assetId });
      set({ 
        isUploading: false, 
        isProcessing: true,
        currentScan: { scan_id: result.scan_id, status: 'processing', image_url: '' }
      });
      
      // Start polling automatically
      await get().pollForResult(result.scan_id);
    } catch (error) {
      set({ isUploading: false, error: error.message });
    }
  },
  
  pollForResult: async (scanId) => {
    set({ isProcessing: true });
    try {
      const result = await receiptApi.pollScanResult(scanId);
      
      // Pre-populate edited data from extracted data
      const editedData: Partial<receiptApi.ConfirmTransactionRequest> = {};
      if (result.extracted_data) {
        editedData.amount = result.extracted_data.total_amount;
        editedData.date = result.extracted_data.date;
        editedData.description = result.extracted_data.merchant_name;
        editedData.transaction_type = 2; // Expense
        if (result.extracted_data.suggested_category) {
          editedData.category_id = result.extracted_data.suggested_category.id;
        }
      }
      
      set({ 
        isProcessing: false, 
        currentScan: result,
        editedData,
        error: result.status === 'failed' ? result.error_message : null
      });
    } catch (error) {
      set({ isProcessing: false, error: error.message });
    }
  },
  
  updateEditedData: (data) => {
    set({ editedData: { ...get().editedData, ...data } });
  },
  
  confirmTransaction: async (scanId) => {
    const { editedData } = get();
    const result = await receiptApi.confirmReceipt(scanId, editedData as receiptApi.ConfirmTransactionRequest);
    return result.transaction_id;
  },
  
  reset: () => {
    set({
      currentScan: null,
      isUploading: false,
      isProcessing: false,
      error: null,
      editedData: {},
    });
  },
}));
```

---

## Custom Hooks

### useCamera Hook

```typescript
// hooks/useCamera.ts

import { useState, useCallback } from 'react';
import { Camera } from 'expo-camera'; // or react-native-camera

export const useCamera = () => {
  const [hasPermission, setHasPermission] = useState<boolean | null>(null);
  const [flashMode, setFlashMode] = useState<'off' | 'on' | 'auto'>('auto');
  const cameraRef = useRef<Camera>(null);
  
  const requestPermission = useCallback(async () => {
    const { status } = await Camera.requestCameraPermissionsAsync();
    setHasPermission(status === 'granted');
    return status === 'granted';
  }, []);
  
  const takePicture = useCallback(async () => {
    if (cameraRef.current) {
      const photo = await cameraRef.current.takePictureAsync({
        quality: 0.8,
        base64: false,
        exif: false,
      });
      return photo;
    }
    return null;
  }, []);
  
  const toggleFlash = useCallback(() => {
    setFlashMode(mode => {
      if (mode === 'off') return 'on';
      if (mode === 'on') return 'auto';
      return 'off';
    });
  }, []);
  
  return {
    cameraRef,
    hasPermission,
    flashMode,
    requestPermission,
    takePicture,
    toggleFlash,
  };
};
```

### useImagePicker Hook

```typescript
// hooks/useImagePicker.ts

import * as ImagePicker from 'expo-image-picker';

export const useImagePicker = () => {
  const pickImage = async () => {
    const result = await ImagePicker.launchImageLibraryAsync({
      mediaTypes: ImagePicker.MediaTypeOptions.Images,
      allowsEditing: true,
      quality: 0.8,
    });
    
    if (!result.canceled) {
      return result.assets[0];
    }
    return null;
  };
  
  return { pickImage };
};
```

---

## UI Components

### Confidence Indicator

```typescript
// components/receipt/ConfidenceIndicator.tsx

import React from 'react';
import { View, Text, StyleSheet } from 'react-native';

interface Props {
  score: number;
}

export const ConfidenceIndicator: React.FC<Props> = ({ score }) => {
  const getColor = () => {
    if (score >= 80) return '#22C55E'; // Green
    if (score >= 50) return '#F59E0B'; // Yellow
    return '#EF4444'; // Red
  };
  
  const getLabel = () => {
    if (score >= 80) return 'High confidence';
    if (score >= 50) return 'Medium confidence';
    return 'Low confidence';
  };
  
  return (
    <View style={styles.container}>
      <View style={styles.barBackground}>
        <View 
          style={[
            styles.barFill, 
            { width: `${score}%`, backgroundColor: getColor() }
          ]} 
        />
      </View>
      <Text style={[styles.label, { color: getColor() }]}>
        {score.toFixed(0)}% • {getLabel()}
      </Text>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    marginVertical: 8,
  },
  barBackground: {
    height: 6,
    backgroundColor: '#E5E7EB',
    borderRadius: 3,
    overflow: 'hidden',
  },
  barFill: {
    height: '100%',
    borderRadius: 3,
  },
  label: {
    fontSize: 12,
    marginTop: 4,
  },
});
```

### Editable Field

```typescript
// components/receipt/EditableField.tsx

import React, { useState } from 'react';
import { View, Text, TextInput, TouchableOpacity, StyleSheet } from 'react-native';
import { Ionicons } from '@expo/vector-icons';

interface Props {
  label: string;
  value: string;
  onChange: (value: string) => void;
  required?: boolean;
  keyboardType?: 'default' | 'numeric' | 'decimal-pad';
  rightIcon?: React.ReactNode;
  onRightIconPress?: () => void;
  suggested?: boolean;
}

export const EditableField: React.FC<Props> = ({
  label,
  value,
  onChange,
  required,
  keyboardType = 'default',
  rightIcon,
  onRightIconPress,
  suggested,
}) => {
  const [isFocused, setIsFocused] = useState(false);
  
  return (
    <View style={styles.container}>
      <View style={styles.labelRow}>
        <Text style={styles.label}>
          {label} {required && <Text style={styles.required}>*</Text>}
        </Text>
        {suggested && (
          <View style={styles.suggestedBadge}>
            <Text style={styles.suggestedText}>Suggested</Text>
          </View>
        )}
      </View>
      
      <View style={[styles.inputContainer, isFocused && styles.inputFocused]}>
        <TextInput
          style={styles.input}
          value={value}
          onChangeText={onChange}
          onFocus={() => setIsFocused(true)}
          onBlur={() => setIsFocused(false)}
          keyboardType={keyboardType}
        />
        
        {rightIcon ? (
          <TouchableOpacity onPress={onRightIconPress} style={styles.iconButton}>
            {rightIcon}
          </TouchableOpacity>
        ) : (
          <Ionicons name="pencil" size={18} color="#9CA3AF" style={styles.editIcon} />
        )}
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    marginBottom: 16,
  },
  labelRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 6,
  },
  label: {
    fontSize: 14,
    color: '#6B7280',
  },
  required: {
    color: '#EF4444',
  },
  suggestedBadge: {
    backgroundColor: '#DBEAFE',
    paddingHorizontal: 8,
    paddingVertical: 2,
    borderRadius: 4,
    marginLeft: 8,
  },
  suggestedText: {
    fontSize: 10,
    color: '#2563EB',
    fontWeight: '600',
  },
  inputContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    borderWidth: 1,
    borderColor: '#E5E7EB',
    borderRadius: 8,
    backgroundColor: '#F9FAFB',
  },
  inputFocused: {
    borderColor: '#3B82F6',
    backgroundColor: '#FFFFFF',
  },
  input: {
    flex: 1,
    padding: 12,
    fontSize: 16,
    color: '#111827',
  },
  editIcon: {
    marginRight: 12,
  },
  iconButton: {
    padding: 12,
  },
});
```

### Processing Animation

```typescript
// components/receipt/ProcessingAnimation.tsx

import React, { useEffect, useRef } from 'react';
import { View, Text, Animated, StyleSheet } from 'react-native';
import LottieView from 'lottie-react-native';

interface Props {
  progress?: number;
}

export const ProcessingAnimation: React.FC<Props> = ({ progress }) => {
  const scanLine = useRef(new Animated.Value(0)).current;
  
  useEffect(() => {
    const animation = Animated.loop(
      Animated.sequence([
        Animated.timing(scanLine, {
          toValue: 1,
          duration: 1500,
          useNativeDriver: true,
        }),
        Animated.timing(scanLine, {
          toValue: 0,
          duration: 1500,
          useNativeDriver: true,
        }),
      ])
    );
    animation.start();
    return () => animation.stop();
  }, []);
  
  const translateY = scanLine.interpolate({
    inputRange: [0, 1],
    outputRange: [0, 150],
  });
  
  return (
    <View style={styles.container}>
      <View style={styles.documentContainer}>
        <View style={styles.document}>
          <View style={styles.documentLine} />
          <View style={styles.documentLine} />
          <View style={styles.documentLine} />
          <View style={[styles.documentLine, { width: '60%' }]} />
          
          <Animated.View 
            style={[
              styles.scanLine,
              { transform: [{ translateY }] }
            ]} 
          />
        </View>
      </View>
      
      <Text style={styles.title}>Scanning...</Text>
      <Text style={styles.subtitle}>Reading receipt details</Text>
      
      {progress !== undefined && (
        <View style={styles.progressContainer}>
          <View style={styles.progressBar}>
            <View style={[styles.progressFill, { width: `${progress}%` }]} />
          </View>
          <Text style={styles.progressText}>{progress}%</Text>
        </View>
      )}
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    alignItems: 'center',
    justifyContent: 'center',
    padding: 40,
  },
  documentContainer: {
    width: 120,
    height: 160,
    marginBottom: 24,
  },
  document: {
    flex: 1,
    backgroundColor: '#FFFFFF',
    borderRadius: 8,
    padding: 16,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.1,
    shadowRadius: 8,
    elevation: 4,
    overflow: 'hidden',
  },
  documentLine: {
    height: 8,
    backgroundColor: '#E5E7EB',
    borderRadius: 4,
    marginBottom: 12,
  },
  scanLine: {
    position: 'absolute',
    left: 0,
    right: 0,
    height: 2,
    backgroundColor: '#3B82F6',
    shadowColor: '#3B82F6',
    shadowOffset: { width: 0, height: 0 },
    shadowOpacity: 0.8,
    shadowRadius: 4,
  },
  title: {
    fontSize: 20,
    fontWeight: '600',
    color: '#111827',
    marginBottom: 4,
  },
  subtitle: {
    fontSize: 14,
    color: '#6B7280',
    marginBottom: 24,
  },
  progressContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    width: '80%',
  },
  progressBar: {
    flex: 1,
    height: 4,
    backgroundColor: '#E5E7EB',
    borderRadius: 2,
    overflow: 'hidden',
  },
  progressFill: {
    height: '100%',
    backgroundColor: '#3B82F6',
  },
  progressText: {
    marginLeft: 12,
    fontSize: 14,
    color: '#6B7280',
    minWidth: 40,
  },
});
```

---

## Navigation

### Navigation Setup

```typescript
// navigation/ReceiptNavigator.tsx

import { createStackNavigator } from '@react-navigation/stack';
import {
  ReceiptCaptureScreen,
  ImagePreviewScreen,
  ProcessingScreen,
  ReviewEditScreen,
  ItemsDetailScreen,
  SuccessScreen,
  ScanHistoryScreen,
} from '../screens/receipt';

export type ReceiptStackParamList = {
  ReceiptCapture: undefined;
  ImagePreview: { imageUri: string };
  Processing: { scanId: number };
  ReviewEdit: { scanId: number };
  ItemsDetail: { items: ReceiptItem[] };
  Success: { transactionId: number; amount: number; merchant: string };
  ScanHistory: undefined;
};

const Stack = createStackNavigator<ReceiptStackParamList>();

export const ReceiptNavigator = () => (
  <Stack.Navigator
    screenOptions={{
      headerShown: false,
      cardStyleInterpolator: CardStyleInterpolators.forHorizontalIOS,
    }}
  >
    <Stack.Screen name="ReceiptCapture" component={ReceiptCaptureScreen} />
    <Stack.Screen name="ImagePreview" component={ImagePreviewScreen} />
    <Stack.Screen name="Processing" component={ProcessingScreen} />
    <Stack.Screen name="ReviewEdit" component={ReviewEditScreen} />
    <Stack.Screen name="ItemsDetail" component={ItemsDetailScreen} />
    <Stack.Screen name="Success" component={SuccessScreen} />
    <Stack.Screen name="ScanHistory" component={ScanHistoryScreen} />
  </Stack.Navigator>
);
```

---

## Entry Points

### FAB (Floating Action Button) Integration

```typescript
// Add receipt scan as option in transaction FAB

const FABOptions = [
  {
    icon: 'add',
    label: 'Add Transaction',
    onPress: () => navigate('AddTransaction'),
  },
  {
    icon: 'camera',
    label: 'Scan Receipt',
    onPress: () => navigate('ReceiptCapture'),
  },
  {
    icon: 'swap-horizontal',
    label: 'Transfer',
    onPress: () => navigate('Transfer'),
  },
];
```

### Dashboard Quick Action

```
┌─────────────────────────────────────────┐
│  Quick Actions                          │
├─────────────────────────────────────────┤
│                                         │
│  ┌──────────┐  ┌──────────┐  ┌───────┐ │
│  │   ➕     │  │   📸     │  │  🔄   │ │
│  │   Add    │  │  Scan    │  │Transfer│ │
│  └──────────┘  └──────────┘  └───────┘ │
│                                         │
└─────────────────────────────────────────┘
```

---

## Error Handling UX

### Network Error

```
┌─────────────────────────────────────────┐
│                                         │
│              📶 ✕                        │
│                                         │
│      No internet connection             │
│                                         │
│  Your receipt has been saved locally.   │
│  It will be scanned automatically       │
│  when you're back online.               │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │         Retry Now               │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │         Enter Manually          │   │
│  └─────────────────────────────────┘   │
│                                         │
└─────────────────────────────────────────┘
```

### OCR Failed

```
┌─────────────────────────────────────────┐
│                                         │
│              ❌                          │
│                                         │
│     Couldn't read receipt               │
│                                         │
│  The image quality might be too low     │
│  or the receipt format is not           │
│  supported.                             │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │         Try Again               │   │
│  └─────────────────────────────────┘   │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │         Enter Manually          │   │
│  └─────────────────────────────────┘   │
│                                         │
│              [ Cancel ]                 │
│                                         │
└─────────────────────────────────────────┘
```

---

## Accessibility

- **VoiceOver/TalkBack**: Announce processing status
- **High Contrast**: Clear visual feedback for confidence levels
- **Large Touch Targets**: 44x44pt minimum for all buttons
- **Haptic Feedback**: On photo capture, scan complete, errors
- **Screen Reader Labels**: All icons and buttons have accessible labels

---

## Platform-Specific Considerations

### iOS
- Use UIImagePickerController or PHPicker
- Request camera permission with clear purpose string
- Handle photo library permission separately

### Android
- Request CAMERA and READ_EXTERNAL_STORAGE permissions
- Handle Android 11+ scoped storage
- Consider using CameraX for modern implementation

### Web (PWA)
- Use navigator.mediaDevices.getUserMedia
- Handle permission denial gracefully
- Provide file upload as primary option (camera as secondary)

---

## Testing Checklist

- [ ] Camera permission flow (granted/denied)
- [ ] Gallery picker works correctly
- [ ] Image upload with poor network
- [ ] Processing timeout handling
- [ ] Edit and correct extracted data
- [ ] Category selection with suggested highlight
- [ ] Successful transaction creation
- [ ] Error states display correctly
- [ ] Accessibility with screen reader
- [ ] Different receipt types (retail, restaurant, etc.)
