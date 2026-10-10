export interface CityBackdropTheme {
  sky: string;
  glass: string;
  distantGlass: string;
  floorBands: string;
  windows: string;
  crown: string;
}

// Presentation rules can map a future day/night cycle to these render parameters.
// Theme colors and transient lighting state remain independent of city geometry.
export interface CityBackdropState {
  skyBrightness: number;
  windowGlow: number;
}

export const eveningCityTheme: Readonly<CityBackdropTheme> = {
  sky: '#172b43',
  glass: '#1c394a',
  distantGlass: '#294459',
  floorBands: '#496d7d',
  windows: '#a4dce8',
  crown: '#e1b778',
};

export const initialCityState: Readonly<CityBackdropState> = {
  skyBrightness: 1,
  windowGlow: 0.75,
};
