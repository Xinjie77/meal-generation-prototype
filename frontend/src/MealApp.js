import React, { useState } from "react";
import axios from "axios";
import MealTable from "./MealTable";
import MealForm from "./MealForm";
import "./Meal.css";

const MealApp = () => {
  const [meals, setMeals] = useState([]);
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState({
    dietPreference: {
      Mediterranean: false,
      Vegetarian: false,
      Vegan: false,
      Ketogenic: false,
      Nordic: false,
      "Raw food": false,
      Carnivore: false,
      "Low sodium": false,
    },
    purpose: "Maintain Weight",
    fitnessLevel: "Beginner",
    height: 180,
    weight: 75,
    gender: "Male",
    mealsPerDay: 3,
    snacksPerDay: 1,
    allergies: "none",
  });

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData({
      ...formData,
      [name]: value,
    });
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    setLoading(true);

    // Convert height to cm and weight to kg if using imperial units
    let height = parseFloat(formData.height);
    let weight = parseFloat(formData.weight);
    if (formData.heightFeet || formData.heightInches || formData.weightPounds) {
      height =
        parseFloat(formData.heightFeet || 0) * 30.48 +
        parseFloat(formData.heightInches || 0) * 2.54;
      weight = parseFloat(formData.weightPounds) / 2.20462;
    }

    const convertedData = {
      ...formData,
      height: height,
      weight: weight,
      mealsPerDay: parseInt(formData.mealsPerDay,10),
      snacksPerDay: parseInt(formData.snacksPerDay,10),
    };

    console.log("Form data:", convertedData);
    try {
      const response = await axios.post(
        "http://localhost:8080/get-meal-data",
        convertedData,
        {
          headers: {
            "Content-Type": "application/json",
          },
        }
      );
      console.log("Response data:", response.data);
      setMeals(response.data.meals);
    } catch (error) {
      console.error("Error fetching meals:", error);
    } finally {
      setLoading(false);
    }
  };

  const updateMeal = (day, oldMealName, newMeal) => {
    console.log(
      "Updating meal for day:",
      day,
      "with new meal:",
      newMeal
    ); // Log the update details
    setMeals((prevMeals) =>
      prevMeals.map((meal) =>
        meal.day === day && meal.meal === oldMealName
          ? newMeal
          : meal
      )
    );
  };

  return (
    <div className="App">
      <h1>Meal Schedule</h1>
      <MealForm
        formData={formData}
        handleChange={handleChange}
        handleSubmit={handleSubmit}
        loading={loading}
      />
      <MealTable
        meals={meals}
        formData={formData}
        updateMeal={updateMeal}
      />
    </div>
  );
};

export default MealApp;
