import React, { useState } from "react";
import "./Meal.css";

const MealForm = ({ formData, handleChange, handleSubmit, loading }) => {
  const Preferences = [
    "Mediterranean", "Vegetarian", "Vegan", "Ketogenic", "Nordic", "Raw food", "Carnivore", "Low sodium"
  ];

  const purposes = ["Lose Weight", "Gain Muscle", "Maintain Weight"];
  const fitnessLevel = ["Beginner", "I do sport from time to time", "I do sport regularly", "Advanced"];
  const [useMetric, setUseMetric] = useState(true);

  const handleDietPreferenceChange = (event) => {
    const { value, checked } = event.target;
    handleChange({
      target: {
        name: "dietPreference",
        value: {
          ...formData.dietPreference,
          [value]: checked,
        },
      },
    });
  };

  const handleUnitChange = (event) => {
    setUseMetric(event.target.value === "metric");
  };

  const handleInputChange = (event) => {
    const { name, value } = event.target;
    handleChange({
      target: {
        name,
        value: value,
      },
    });
  };

  return (
    <form onSubmit={handleSubmit} className="form-container">
      <div className="form-group">
        <label>Diet Preferences: </label>
        {Preferences.map((preference) => (
          <div key={preference}>
            <label>
              <input
                type="checkbox"
                value={preference}
                checked={formData.dietPreference[preference] || false}
                onChange={handleDietPreferenceChange}
              />
              {preference}
            </label>
          </div>
        ))}
      </div>
      <div className="form-group">
        <label>Purpose:</label>
        <select
          name="purpose"
          value={formData.purpose}
          onChange={handleInputChange}
        >
          {purposes.map((dietpurpose) => (
            <option key={dietpurpose} value={dietpurpose}>
              {dietpurpose.charAt(0).toUpperCase() + dietpurpose.slice(1)}
            </option>
          ))}
        </select>
      </div>
      <div className="form-group">
        <label>Fitness Level:</label>
        <select
          name="fitnessLevel"
          value={formData.fitnessLevel}
          onChange={handleInputChange}
        >
          {fitnessLevel.map((level) => (
            <option key={level} value={level}>
              {level.charAt(0).toUpperCase() + level.slice(1)}
            </option>
          ))}
        </select>
      </div>
      <div className="form-group">
        <label>Units:</label>
        <select onChange={handleUnitChange}>
          <option value="metric">Metric (cm, kg)</option>
          <option value="imperial">Imperial (ft/in, lbs)</option>
        </select>
      </div>
      {useMetric ? (
        <>
          <div className="form-group">
            <label>Height (cm):</label>
            <input
              type="number"
              name="height"
              value={formData.height}
              onChange={handleInputChange}
            />
          </div>
          <div className="form-group">
            <label>Weight (kg):</label>
            <input
              type="number"
              name="weight"
              value={formData.weight}
              onChange={handleInputChange}
            />
          </div>
        </>
      ) : (
        <>
          <div className="form-group">
            <label>Height:</label>
            <input
              type="number"
              name="heightFeet"
              placeholder="Feet"
              value={formData.heightFeet || ""}
              onChange={handleInputChange}
              style={{ marginRight: "10px", width: "calc(50% - 5px)" }}
            />
            <input
              type="number"
              name="heightInches"
              placeholder="Inches"
              value={formData.heightInches || ""}
              onChange={handleInputChange}
              style={{ width: "calc(50% - 5px)" }}
            />
          </div>
          <div className="form-group">
            <label>Weight (lbs):</label>
            <input
              type="number"
              name="weightPounds"
              value={formData.weightPounds || ""}
              onChange={handleInputChange}
            />
          </div>
        </>
      )}
      <div className="form-group">
        <label>Gender:</label>
        <select
          name="gender"
          value={formData.gender}
          onChange={handleInputChange}
        >
          <option value="Male">Male</option>
          <option value="Female">Female</option>
        </select>
      </div>

      <div className="form-group">
        <label>Meals per Day:</label>
        <input
          type="number"
          name="mealsPerDay"
          value={formData.daysPerDay}
          onChange={handleInputChange}
        />
      </div>
      <div className="form-group">
        <label>Snacks per Day:</label>
        <input
          type="number"
          name="snacksPerDay"
          value={formData.hoursPerDay}
          onChange={handleInputChange}
        />
      </div>
      <div className="form-group">
        <label>Allergies: </label>
        <input
          type="text"
          name="allergies"
          value={formData.allergies}
          onChange={handleInputChange}
        />
      </div>
      <button type="submit" className="submit-button" disabled={loading}>
        {loading ? "Loading..." : "Submit"}
      </button>
    </form>
  );
};

export default MealForm;
